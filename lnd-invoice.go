package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os/user"
	"path"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/macaroons"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/macaroon.v2"
)

type API struct {
	lnrpc.LightningClient
}

func NewClient() API {
	usr, err := user.Current()
	if err != nil {
		log.Println("Cannot get current user:", err)
	}
	tlsCertPath := path.Join(usr.HomeDir, ".lnd/tls.cert")
	macaroonPath := path.Join(usr.HomeDir, ".lnd/invoice.macaroon")

	tlsCreds, err := credentials.NewClientTLSFromFile(tlsCertPath, "")
	if err != nil {
		log.Println("Cannot get node tls credentials", err)
	}

	macaroonBytes, err := ioutil.ReadFile(macaroonPath)
	if err != nil {
		log.Println("Cannot read macaroon file", err)
	}

	mac := &macaroon.Macaroon{}
	if err = mac.UnmarshalBinary(macaroonBytes); err != nil {
		log.Println("Cannot unmarshal macaroon", err)
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(tlsCreds),
		grpc.WithBlock(),
		grpc.WithPerRPCCredentials(macaroons.NewMacaroonCredential(mac)),
	}

	conn, err := grpc.Dial("192.168.11.59:10009", opts...)
	if err != nil {
		log.Println("cannot dial to lnd", err)
	}
	client := lnrpc.NewLightningClient(conn)

	return API{client}

}

func main() {
	c := NewClient()

	invoice, err := c.AddInvoice(context.Background(), &lnrpc.Invoice{
		Memo:  "boli-lightning",
		Value: 250,
	})
	if err != nil {
		log.Println(err)
	}

	fmt.Println("\nPayment Request:", invoice.GetPaymentRequest())
	fmt.Println("Invoice:", invoice)
}
