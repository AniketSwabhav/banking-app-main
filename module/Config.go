package module

import (
	"banking-app-be/app"
	"banking-app-be/model/account"
	"banking-app-be/model/bank"
	banktransaction "banking-app-be/model/bankTransaction"
	"banking-app-be/model/credential"
	"banking-app-be/model/passbook"
	"banking-app-be/model/user"
)

func Configure(appObj *app.App) {
	appObj.Log.Print("============Configuring-Module-Configs==============")

	userModule := user.NewUserModuleConfig(appObj.DB)
	credentialModule := credential.NewCredentialModuleConfig(appObj.DB)
	bankModule := bank.NewBankModuleConfig(appObj.DB)
	banktransactionModule := banktransaction.NewBankTransactionModuleConfig(appObj.DB)
	accountModule := account.NewAccountModuleConfig(appObj.DB)
	passbookModule := passbook.NewPassbookModuleConfig(appObj.DB)

	appObj.MigrateModuleTables([]app.ModuleConfig{userModule, credentialModule, bankModule, banktransactionModule, accountModule, passbookModule})
}
