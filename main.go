package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/digitalocean/godo"
	"github.com/golang/glog"
	"golang.org/x/oauth2"
)

var (
	token       = flag.String("token", "", "digital ocean token")
	dropletName = flag.String("droplet-name", "super-cool-droplet", "droplet name")
	action      = flag.String("action", "get", "get/create/delete a droplet")
)

type TokenSource struct {
	AccessToken string
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

func main() {
	flag.Parse()
	flag.Set("logtostderr", "true")
	client := createClient()
	switch *action {
	case "create":
		if err := createDroplet(client, *dropletName); err != nil {
			glog.Fatal(err)
		}
	case "get":
		if droplet, err := getDroplet(client, *dropletName); err != nil {
			glog.Fatal(err)
		} else {
			glog.Infof("droplet %v", droplet)
		}
	case "delete":
		if err := deleteDroplet(client, *dropletName); err != nil {
			glog.Fatal(err)
		}
	}
}

func createClient() *godo.Client {
	tokenSource := &TokenSource{
		AccessToken: *token,
	}
	oauthClient := oauth2.NewClient(context.Background(), tokenSource)
	return godo.NewClient(oauthClient)
}

func listSSHKeys(client *godo.Client) ([]godo.Key, error) {
	list, _, err := client.Keys.List(context.TODO(), nil)
	return list, err
}

func getDroplet(client *godo.Client, dropletName string) (*godo.Droplet, error) {
	list, _, err := client.Droplets.List(context.TODO(), nil)
	if err != nil {
		return nil, err
	}
	for i := range list {
		if list[i].Name == dropletName {
			return &list[i], nil
		}
	}
	return nil, nil
}

func createDroplet(client *godo.Client, dropletName string) error {
	keys, err := listSSHKeys(client)
	if err != nil {
		return err
	}
	var sshKeys []godo.DropletCreateSSHKey
	for _, key := range keys {
		sshKeys = append(sshKeys, godo.DropletCreateSSHKey{ID: key.ID, Fingerprint: key.Fingerprint})
	}

	createRequest := &godo.DropletCreateRequest{
		Name:   dropletName,
		Region: "nyc3",
		Size:   "s-1vcpu-1gb",
		Image: godo.DropletCreateImage{
			Slug: "ubuntu-18-04-x64",
		},
		SSHKeys: sshKeys,
	}
	droplet, _, err := client.Droplets.Create(context.TODO(), createRequest)

	if err != nil {
		fmt.Printf("Something bad happened: %s\n\n", err)
		return err
	}
	glog.Infof("created droplet %v", droplet)
	return nil
}

func deleteDroplet(client *godo.Client, dropletName string) error {
	droplet, err := getDroplet(client, dropletName)
	if err != nil {
		return err
	}
	if droplet == nil {
		return errors.New("droplet not found")
	}
	_, err = client.Droplets.Delete(context.TODO(), droplet.ID)
	if err == nil {
		glog.Infof("deleted droplet %v", droplet)
	}
	return err
}
