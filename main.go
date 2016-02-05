package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/bluele/slack"
)

const filename = "notified.json"

var token = kingpin.Flag("token", "The slack API token").OverrideDefaultFromEnvar("TOKEN").Required().String()
var timeout = kingpin.Arg("wait", "Wait in minutes before checking for changes").Default("10").Int()
var expireTime = kingpin.Arg("expire", "Age in hours before notifing about the instance").Default("48").Int()
var region = kingpin.Arg("region", "Amazon EC2 region to scan for instances").Default("eu-west-1").String()
var prod = kingpin.Flag("prod", "Running for real?").OverrideDefaultFromEnvar("PROD").Default("false").Bool()

func findUser(list []*slack.User, searchid string) string {
	for _, u := range list {
		if u.Id == searchid {
			return u.Id
		}

		if u.Profile.Email != "" {
			data := strings.Split(u.Profile.Email, "@")
			if data[0] == searchid {
				return u.Id
			}
		}
	}
	return ""
}

func handleShutdown() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("Shutting down.")
		os.Exit(1)
	}()

}
func main() {
	kingpin.Version("1.3")
	kingpin.Parse()

	api := slack.New(*token)

	for {

		users, _ := api.UsersList()

		svc := ec2.New(session.New(), &aws.Config{Region: aws.String(*region)})

		// Call the DescribeInstances Operation
		resp, err := svc.DescribeInstances(nil)
		if err != nil {
			panic(err)
		}

		usersInstances := findLongRunningInstances(resp, users)
		alreadyNotified := getAlreadyNotified()

		for user, instances := range usersInstances {
			msg := ""
			for _, inst := range instances {
				if !contains(alreadyNotified, *inst.InstanceId) {
					msg += fmt.Sprintf("`%v - %v` with id _%v_\n", getTagValue(inst.Tags, "Purpose"), getTagValue(inst.Tags, "Name"), *inst.InstanceId)
					alreadyNotified = append(alreadyNotified, *inst.InstanceId)
				}
			}
			if msg != "" {
				msg := "\nYou have the following instances running in amazon:\n\n" + msg
				msg += "They were all started more then two days ago, are they still needed ?"
				if user == "U02HSGZ3F" || *prod { // ME
					err = api.ChatPostMessage(user, msg, &slack.ChatPostMessageOpt{Username: "Tradeshift AWS Service notifier"})
				}

			}

		}
		saveAlreadyNotified(alreadyNotified)
		time.Sleep(time.Duration(*timeout) * time.Minute)
	}
}

func findLongRunningInstances(resp *ec2.DescribeInstancesOutput, users []*slack.User) map[string][]*ec2.Instance {
	usersInstances := map[string][]*ec2.Instance{}
	// resp has all of the response data, pull out instance IDs:
	fmt.Println("> Number of instances: ", len(resp.Reservations))
	for _, res := range resp.Reservations {
		for _, inst := range res.Instances {
			if time.Since(*inst.LaunchTime) > time.Duration((*expireTime))*time.Hour {
				for _, tag := range inst.Tags {
					if *tag.Key == "Owner" {
						owner := *tag.Value
						userid := findUser(users, *tag.Value)
						if userid != "" {
							usersInstances[userid] = append(usersInstances[userid], inst)
						} else {
							log.Println("Unable to locate user id for ", owner)
						}

					}

				}
			}

		}
	}
	return usersInstances
}

func saveAlreadyNotified(notified []string) {
	b, err := json.Marshal(notified)
	if err != nil {
		log.Print("Unable to marshal list of notified: ", err)
	}
	err = ioutil.WriteFile(filename, b, 0644)
	if err != nil {
		log.Print("Unable to save json file")
	}
}

func getAlreadyNotified() []string {
	file, e := ioutil.ReadFile(filename)
	if e != nil {
		log.Println("File error: ", e)
	}

	var alreadyNotified []string
	json.Unmarshal(file, &alreadyNotified)
	return alreadyNotified
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func getTagValue(tags []*ec2.Tag, key string) string {
	for _, tag := range tags {
		if *tag.Key == key {
			return *tag.Value
		}
	}
	return ""
}
