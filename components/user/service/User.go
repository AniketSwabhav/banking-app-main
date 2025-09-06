package user

import (
	"banking-app-be/components/errors"
	"banking-app-be/components/log"
	"banking-app-be/components/security"
	"banking-app-be/model/credential"
	"banking-app-be/model/user"
	"banking-app-be/module/repository"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

const cost = 10

type UserService struct {
	db         *gorm.DB
	repository repository.Repository
}

func NewUserService(DB *gorm.DB, repo repository.Repository) *UserService {
	return &UserService{
		db:         DB,
		repository: repo,
	}
}

func (service *UserService) CreateAdmin(newUser *user.User) error {

	uow := repository.NewUnitOfWork(service.db, false)
	defer uow.RollBack()

	err := service.doesEmailExists(newUser.Credentials.Email)
	if err != nil {
		return err
	}

	if err = newUser.Validate(); err != nil {
		log.GetLogger().Error(err.Error())
		uow.RollBack()
		return err
	}

	if err := newUser.Credentials.Validate(); err != nil {
		return err
	}

	*newUser.IsAdmin = true

	hashedPassword, err := hashPassword(newUser.Credentials.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	newUser.Credentials.Password = string(hashedPassword)

	err = uow.DB.Create(newUser).Error
	if err != nil {
		return errors.NewDatabaseError("Failed to create user")
	}

	uow.Commit()
	return nil
}

func (service *UserService) CreateUser(newUser *user.User) error {

	uow := repository.NewUnitOfWork(service.db, false)
	defer uow.RollBack()

	err := service.doesEmailExists(newUser.Credentials.Email)
	if err != nil {
		return err
	}

	if err = newUser.Validate(); err != nil {
		log.GetLogger().Error(err.Error())
		uow.RollBack()
		return err
	}

	if err := newUser.Credentials.Validate(); err != nil {
		return err
	}

	hashedPassword, err := hashPassword(newUser.Credentials.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	newUser.Credentials.Password = string(hashedPassword)

	err = uow.DB.Create(newUser).Error
	if err != nil {
		return errors.NewDatabaseError("Failed to create user")
	}

	uow.Commit()
	return nil
}

func (service *UserService) Login(userCredential *credential.Credential, claim *security.Claims) error {

	uow := repository.NewUnitOfWork(service.db, false)
	defer uow.RollBack()

	exists, err := repository.DoesEmailExist(service.db, userCredential.Email, credential.Credential{},
		repository.Filter("`email` = ?", userCredential.Email))
	if err != nil {
		return errors.NewDatabaseError("Error checking if email exists")
	}
	if !exists {
		return errors.NewNotFoundError("Email not found")
	}

	foundCredential := credential.Credential{}
	err = uow.DB.Where("email = ?", userCredential.Email).First(&foundCredential).Error
	if err != nil {
		return errors.NewDatabaseError("Could not retrieve credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(foundCredential.Password), []byte(userCredential.Password))
	if err != nil {
		return errors.NewInValidPasswordError("Incorrect password")
	}

	foundUser := user.User{}
	err = uow.DB.Preload("Credentials").
		Where("id = ?", foundCredential.UserID).First(&foundUser).Error

	if err != nil {
		return errors.NewDatabaseError("Could not retrieve user")
	}

	*claim = security.Claims{
		UserID:   foundUser.ID,
		IsAdmin:  *foundUser.IsAdmin,
		IsActive: *foundUser.IsActive,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(30 * time.Hour).Unix(),
		},
	}

	uow.Commit()
	return nil
}

func (service *UserService) GetAllUsers(allUsers *[]user.UserDTO, totalCount *int, limit, offset int) error {

	uow := repository.NewUnitOfWork(service.db, true)
	defer uow.RollBack()

	err := service.repository.GetAll(uow, allUsers, repository.PreloadAssociations([]string{"Credentials", "Accounts"}), repository.Paginate(limit, offset, totalCount))
	if err != nil {
		return err
	}

	err = service.repository.GetCount(uow, allUsers, totalCount)
	if err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (service *UserService) GetUserByID(targetUser *user.UserDTO) error {

	uow := repository.NewUnitOfWork(service.db, true)
	defer uow.RollBack()

	err := service.repository.GetRecordByID(uow, targetUser.ID, targetUser, repository.PreloadAssociations([]string{"Credentials", "Accounts", "Accounts.Bank"}))
	if err != nil {
		return err
	}

	return nil
}

func (service *UserService) UpdateUser(userToUpdate *user.User) error {
	err := service.doesUserExist(userToUpdate.ID)
	if err != nil {
		return err
	}

	uow := repository.NewUnitOfWork(service.db, false)
	defer uow.RollBack()

	existingUser := user.User{}
	if err := service.repository.GetRecordByID(uow, userToUpdate.ID, &existingUser); err != nil {
		return err
	}

	if err := service.repository.UpdateWithMap(uow, &user.User{}, map[string]interface{}{
		"first_name": userToUpdate.FirstName,
		"last_name":  userToUpdate.LastName,
		"phone_no":   userToUpdate.PhoneNo,
		"is_admin":   userToUpdate.IsAdmin,
		"is_active":  userToUpdate.IsActive,
		"updated_by": userToUpdate.UpdatedBy,
		"updated_at": time.Now(),
	}, repository.Filter("id = ?", userToUpdate.ID)); err != nil {
		uow.RollBack()
		return err
	}

	if userToUpdate.Credentials != nil {
		cred := userToUpdate.Credentials

		if err := cred.Validate(); err != nil {
			uow.RollBack()
			return err
		}

		var updateData = map[string]interface{}{
			"email":      cred.Email,
			"updated_by": userToUpdate.UpdatedBy,
			"updated_at": time.Now(),
		}

		if cred.Password != "" {
			hashedPassword, err := hashPassword(cred.Password)
			if err != nil {
				uow.RollBack()
				return errors.NewValidationError("Failed to hash password")
			}
			updateData["password"] = string(hashedPassword)
		}

		if err := service.repository.UpdateWithMap(uow, &credential.Credential{}, updateData,
			repository.Filter("user_id = ?", userToUpdate.ID)); err != nil {
			uow.RollBack()
			return err
		}
	}

	uow.Commit()
	return nil
}

func (service *UserService) NormalUpdate(userToUpdate *user.User) error {

	err := service.doesUserExist(userToUpdate.ID)
	if err != nil {
		return err
	}

	uow := repository.NewUnitOfWork(service.db, false)
	defer uow.RollBack()

	fmt.Printf("userToupdate %+v", userToUpdate)
	if err := service.repository.Update(uow, userToUpdate); err != nil {
		uow.RollBack()
		return errors.NewDatabaseError("Unable to update user record")
	}

	if userToUpdate.Credentials != nil {
		cred := userToUpdate.Credentials

		if err := cred.Validate(); err != nil {
			uow.RollBack()
			return err
		}

		if cred.Password != "" {
			hashedPassword, err := hashPassword(cred.Password)
			if err != nil {
				uow.RollBack()
				return errors.NewValidationError("Failed to hash password")
			}
			userToUpdate.Credentials.Password = string(hashedPassword)
		}

		if err := service.repository.Update(uow, cred, repository.Filter("user_id = ?", userToUpdate.ID)); err != nil {
			uow.RollBack()
			errors.NewDatabaseError("Unable to update the user credentials")
		}
	}

	uow.Commit()
	return nil
}

func (service *UserService) Delete(userToDelete *user.User) error {

	err := service.doesUserExist(userToDelete.ID)
	if err != nil {
		return err
	}

	uow := repository.NewUnitOfWork(service.db, false)
	defer uow.RollBack()

	if err := service.repository.UpdateWithMap(uow, userToDelete, map[string]interface{}{
		"deleted_at": time.Now(),
		"deleted_by": userToDelete.DeletedBy,
	},
		repository.Filter("`id`=?", userToDelete.ID)); err != nil {
		uow.RollBack()
		return err
	}

	if err := service.repository.UpdateWithMap(uow, &credential.Credential{}, map[string]interface{}{
		"deleted_at": time.Now(),
		"deleted_by": userToDelete.DeletedBy,
	}, repository.Filter("user_id = ?", userToDelete.ID)); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

//==================================================================================================================================

func (service *UserService) doesEmailExists(Email string) error {
	exists, _ := repository.DoesEmailExist(service.db, Email, credential.Credential{},
		repository.Filter("`email` = ?", Email))
	if exists {
		return errors.NewValidationError("Email is already registered")
	}
	return nil
}

func (service *UserService) doesUserExist(ID uuid.UUID) error {
	exists, err := repository.DoesRecordExistForUser(service.db, ID, user.User{},
		repository.Filter("`id` = ?", ID))
	if !exists || err != nil {
		return errors.NewValidationError("User ID is Invalid")
	}
	return nil
}

func hashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), cost)
}
