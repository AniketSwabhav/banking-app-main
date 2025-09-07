package service

import (
	"banking-app-be/components/errors"
	"banking-app-be/components/log"
	"banking-app-be/model/bank"
	banktransaction "banking-app-be/model/bankTransaction"
	"banking-app-be/model/user"
	"banking-app-be/module/repository"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

type BankService struct {
	db         *gorm.DB
	repository repository.Repository
}

func NewBankService(DB *gorm.DB, repo repository.Repository) *BankService {
	return &BankService{
		db:         DB,
		repository: repo,
	}
}

func (service *BankService) CreateBank(newBank *bank.Bank) error {
	uow := repository.NewUnitOfWork(service.db, false)
	defer uow.RollBack()

	if err := newBank.Validate(); err != nil {
		log.GetLogger().Error(err.Error())
		return err
	}

	if err := uow.DB.Create(newBank).Error; err != nil {
		return errors.NewDatabaseError("Failed to create bank")
	}

	uow.Commit()
	return nil
}

func (service *BankService) GetAllBanks(allBanks *[]bank.BankDTO, totalCount *int, limit, offset int) error {

	uow := repository.NewUnitOfWork(service.db, true)
	defer uow.RollBack()

	//repository.PreloadAssociations([]string{"Accounts", "BankTransactions"}),
	err := service.repository.GetAll(uow, allBanks, repository.Paginate(limit, offset, totalCount))
	if err != nil {
		return err
	}

	err = service.repository.GetCount(uow, allBanks, totalCount)
	if err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (service *BankService) GetBankByID(targetBank *bank.BankDTO) error {

	uow := repository.NewUnitOfWork(service.db, true)
	defer uow.RollBack()

	//repository.PreloadAssociations([]string{"Accounts", "bankTransactions"})
	err := service.repository.GetRecordByID(uow, targetBank.ID, targetBank, repository.PreloadAssociations([]string{"Accounts", "Accounts.User"}))
	if err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (service *BankService) UpdateBank(bankToUpdate *bank.Bank) error {

	err := service.doesBankExist(bankToUpdate.ID)
	if err != nil {
		return err
	}

	uow := repository.NewUnitOfWork(service.db, false)
	defer uow.RollBack()

	// existingBank := bank.Bank{}
	// if err := service.repository.GetRecordByID(uow, bankToUpdate.ID, &existingBank); err != nil {
	// 	return err
	// }

	// updateData := map[string]interface{}{
	// 	"full_name":  bankToUpdate.FullName,
	// 	"updated_by": bankToUpdate.UpdatedBy,
	// 	"updated_at": time.Now(),
	// }
	// if bankToUpdate.FullName != "" && bankToUpdate.FullName != existingBank.FullName {
	// 	updateData["abbreviation"] = bank.GetAbbreviation(bankToUpdate.FullName)
	// }
	// if err := service.repository.UpdateWithMap(uow, &bank.Bank{}, updateData, repository.Filter("id = ?", bankToUpdate.ID)); err != nil {
	// 	return err
	// }

	if err := service.repository.Update(uow, bankToUpdate); err != nil {
		uow.RollBack()
		return errors.NewDatabaseError("Unable to update bank record")
	}

	uow.Commit()
	return nil
}

func (service *BankService) DeleteBank(bankToDelete *bank.Bank) error {
	uow := repository.NewUnitOfWork(service.db, false)
	defer uow.RollBack()

	if err := service.repository.UpdateWithMap(uow, bankToDelete, map[string]interface{}{
		"deleted_at": time.Now(),
		"deleted_by": bankToDelete.DeletedBy,
		"is_active":  false,
	},
		repository.Filter("`id`=?", bankToDelete.ID)); err != nil {
		uow.RollBack()
		return err
	}

	uow.Commit()
	return nil
}

func (service *BankService) Settlement(userId uuid.UUID, ledger *[]banktransaction.BankTransactionDTO, totalCount *int) error {
	uow := repository.NewUnitOfWork(service.db, true)
	defer uow.RollBack()

	// Verify user is admin and active
	adminUser := user.User{}
	if err := service.repository.GetRecordByID(uow, userId, &adminUser); err != nil {
		return err
	}
	if adminUser.IsAdmin == nil || !*adminUser.IsAdmin || adminUser.IsActive == nil || !*adminUser.IsActive {
		return errors.NewValidationError("Only active admin users can access settlement records")
	}

	// Fetch all transactions
	allTransactions := []banktransaction.BankTransaction{}
	if err := service.repository.GetAll(uow, &allTransactions); err != nil {
		return errors.NewDatabaseError("Unable to fetch bank transaction entries")
	}

	// Build pairwise ledger
	pairwise := make(map[uuid.UUID]map[uuid.UUID]float32)
	for _, tx := range allTransactions {
		if _, exists := pairwise[tx.SenderBankID]; !exists {
			pairwise[tx.SenderBankID] = make(map[uuid.UUID]float32)
		}
		pairwise[tx.SenderBankID][tx.ReceiverBankID] += tx.Amount
	}

	processed := make(map[string]bool)
	netSettlements := []banktransaction.BankTransactionDTO{}

	for fromBank, toMap := range pairwise {
		for toBank, amountFromTo := range toMap {
			pairKey := fromBank.String() + "->" + toBank.String()
			reversePairKey := toBank.String() + "->" + fromBank.String()

			if processed[pairKey] || processed[reversePairKey] {
				continue
			}

			amountToFrom := pairwise[toBank][fromBank]
			net := amountFromTo - amountToFrom

			if net > 0 {
				netSettlements = append(netSettlements, banktransaction.BankTransactionDTO{
					SenderBankID:   toBank,
					ReceiverBankID: fromBank,
					Amount:         net,
				})
			} else if net < 0 {
				netSettlements = append(netSettlements, banktransaction.BankTransactionDTO{
					SenderBankID:   fromBank,
					ReceiverBankID: toBank,
					Amount:         -net,
				})
			}

			processed[pairKey] = true
			processed[reversePairKey] = true
		}
	}

	// Load bank names for each settlement
	for i, tx := range netSettlements {
		// Load Sender Bank Name
		senderBank := banktransaction.SenderBankName{}
		if err := service.repository.GetRecordByID(uow, tx.SenderBankID, &senderBank); err != nil {
			return fmt.Errorf("failed to load sender bank name: %w", err)
		}
		tx.SenderBank = senderBank

		// Load Receiver Bank Name
		receiverBank := banktransaction.ReceiverBankName{}
		if err := service.repository.GetRecordByID(uow, tx.ReceiverBankID, &receiverBank); err != nil {
			return fmt.Errorf("failed to load receiver bank name: %w", err)
		}
		tx.ReceiverBank = receiverBank

		netSettlements[i] = tx
	}

	*ledger = netSettlements
	*totalCount = len(netSettlements)
	uow.Commit()
	return nil
}

//=======================================================================================

func (service *BankService) doesBankExist(ID uuid.UUID) error {
	exists, err := repository.DoesRecordExistForUser(service.db, ID, bank.Bank{},
		repository.Filter("`id` = ?", ID))
	if !exists || err != nil {
		return errors.NewValidationError("User ID is Invalid")
	}
	return nil
}
