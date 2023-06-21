package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/network"
	"github.com/stellar/go/txnbuild"
)

func createAccount() (map[string]string, error) {
	pair, err := keypair.Random()
	if err != nil {
		return nil, err
	}

	// fmt.Println("Address ", pair.Address())
	// fmt.Println("Seed: ", pair.Seed())

	print("")

	account := map[string]string{
		"public_key": pair.Address(),
		"secret_key": pair.Seed(),
	}

	return account, nil
}

func fundAccount(ddresses [2]string) {

	for _, address := range ddresses {
		resp, err := http.Get("https://friendbot.stellar.org/?addr=" + address)

		if err != nil {
			log.Fatal(err)
		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
			log.Fatal(string(body))
		}

	}
}

func fetchWalletBalances(addresses [2]string) {
	for _, address := range addresses {
		request := horizonclient.AccountRequest{AccountID: address}

		account, err := horizonclient.DefaultTestNetClient.AccountDetail(request)

		if err != nil {
			log.Fatal(err)
		}

		log.Println("Balances for account:", address)

		for _, balance := range account.Balances {
			log.Println(balance)
		}

	}
}

func sendLumens(amount string, sourceSeed string, destinationAddress string) {

	client := horizonclient.DefaultTestNetClient

	// Make sure destination account exists
	destAccountRequest := horizonclient.AccountRequest{AccountID: destinationAddress}
	destinationAccount, err := client.AccountDetail(destAccountRequest)
	if err != nil {
		panic(err)
	}

	fmt.Println("Destination Account", destinationAccount)

	// Load the source account
	sourceKP := keypair.MustParseFull(sourceSeed)
	sourceAccountRequest := horizonclient.AccountRequest{AccountID: sourceKP.Address()}
	sourceAccount, err := client.AccountDetail(sourceAccountRequest)
	if err != nil {
		panic(err)
	}

	// Build transaction
	tx, err := txnbuild.NewTransaction(
		txnbuild.TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			BaseFee:              txnbuild.MinBaseFee,
			Preconditions: txnbuild.Preconditions{
				TimeBounds: txnbuild.NewInfiniteTimeout(),
			},
			Operations: []txnbuild.Operation{
				&txnbuild.Payment{
					Destination: destinationAddress,
					Amount:      amount,
					Asset:       txnbuild.NativeAsset{},
				},
			},
		},
	)

	if err != nil {
		panic(err)
	}

	// Sign the transaction to prove you are actually the person sending it.
	tx, err = tx.Sign(network.TestNetworkPassphrase, sourceKP)
	if err != nil {
		panic(err)
	}

	// And finally, send it off to Stellar!
	resp, err := horizonclient.DefaultTestNetClient.SubmitTransaction(tx)
	if err != nil {
		panic(err)
	}

	fmt.Println("Successful Transaction:")
	fmt.Println("Ledger:", resp.Ledger)
	fmt.Println("Hash:", resp.Hash)

}

func main() {
	deAaddress, err := createAccount()

	if err != nil {
		log.Fatal(err)
	}

	destinationSeed := deAaddress["secret_key"]
	destinationAddress := deAaddress["public_key"]

	fmt.Println("Destination Address: ", destinationAddress)
	fmt.Println("Destination Seed: ", destinationSeed)
	fmt.Println("")

	recAddress, err := createAccount()

	if err != nil {
		log.Fatal(err)
	}

	receiverSeed := recAddress["secret_key"]
	receiverAddress := recAddress["public_key"]

	fmt.Println("Receiver Address: ", receiverAddress)
	fmt.Println("Receiver Seed: ", receiverSeed)
	fmt.Println("")

	addresses := [2]string{destinationAddress, receiverAddress}

	fundAccount(addresses)
	fmt.Println("")
	fetchWalletBalances(addresses)
	fmt.Println("")
	sendLumens("10", destinationSeed, receiverAddress)

}
