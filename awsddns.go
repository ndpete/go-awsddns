package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

type DDNSService struct {
	ZoneID     string
	DomainName string
	IP         string
	client     *route53.Client
}

func main() {
	zoneID := flag.String("zoneid", "", "ZoneID for hosted zone in format /hostedzone/ZONEID")
	domain := flag.String("domain", "", "domain name to update ending with . ie test.example.com. ")
	flag.Parse()
	if *zoneID == "" {
		fmt.Println("Must specifiy zoneid")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *domain == "" {
		fmt.Println("Must specifiy domain")
		flag.PrintDefaults()
		os.Exit(1)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	client := route53.NewFromConfig(cfg)
	ddnsService := &DDNSService{
		ZoneID:     *zoneID,
		DomainName: *domain,
		IP:         getExternalIP(),
		client:     client,
	}

	if ddnsService.checkRecordSet() {
		ddnsService.updateRecordSet()
	}
	log.Printf("Finished")

}

func getExternalIP() string {
	log.Printf("Getting Current IP")
	resp, err := http.Get("https://checkip.amazonaws.com")
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	body = bytes.TrimSpace(body)
	return checkIPAddress(string(body))
}

func checkIPAddress(ip string) string {
	if net.ParseIP(ip) == nil {
		log.Fatalf("IP Address: %s - Invalid\n", ip)
		return ""
	} else {
		log.Printf("IP Address: %s - Valid\n", ip)
		return ip
	}
}

func (d *DDNSService) checkRecordSet() bool {
	log.Printf("Checking Current Record Set")
	params := &route53.ListResourceRecordSetsInput{
		HostedZoneId: &d.ZoneID,
	}
	resp, err := d.client.ListResourceRecordSets(context.TODO(), params)
	if err != nil {
		log.Fatalln(err)
	}
	for _, record := range resp.ResourceRecordSets {
		if aws.ToString(record.Name) == d.DomainName {
			log.Printf("Found record set: %s", aws.ToString(record.Name))
			for _, r := range record.ResourceRecords {
				if aws.ToString(r.Value) != d.IP {
					log.Printf("IP doesn't match on record: %s", aws.ToString(r.Value))
					return true
				}
			}
		} else {
			log.Printf("Record not found")
			return true
		}
	}
	return false
}

func (d *DDNSService) updateRecordSet() {
	record := []types.ResourceRecord{{Value: aws.String(d.IP)}}
	change := types.Change{
		Action: types.ChangeActionUpsert,
		ResourceRecordSet: &types.ResourceRecordSet{
			Name:            &d.DomainName,
			Type:            types.RRTypeA,
			TTL:             aws.Int64(300),
			ResourceRecords: record,
		},
	}
	changes := []types.Change{change}
	changeResouceInput := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: &d.ZoneID,
		ChangeBatch: &types.ChangeBatch{
			Changes: changes,
		},
	}
	log.Printf("Updating DNS Record")
	_, err := d.client.ChangeResourceRecordSets(context.TODO(), changeResouceInput)
	if err != nil {
		log.Fatalf(err.Error())
	}
}
