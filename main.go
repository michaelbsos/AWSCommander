package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

var (
	region    string
	commandId string
	quiet     bool
	html      bool
)

func main() {
	flag.StringVar(&region, "region", "ap-southeast-2", "the AWS region to operate in")
	flag.StringVar(&commandId, "commandId", "", "the command_id for a previously ran command")
	flag.BoolVar(&html, "html", false, "output as HTML")
	flag.BoolVar(&quiet, "quiet", false, "be quiet")
	flag.Parse()

	// -command is required
	if commandId == "" {
		flag.Usage()
		os.Exit(1)
	}

	if html {
		fmt.Printf("<h1>%s - %s</h1>", region, commandId)
	}

	if !quiet {
		fmt.Printf("Looking for command '%s' in '%s'\n", commandId, region)
	}

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Create SSM client
	client := ssm.NewFromConfig(cfg)
	cmds, err := client.ListCommands(context.TODO(), &ssm.ListCommandsInput{
		CommandId: &commandId,
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, cmd := range cmds.Commands {
		// Output UL of instance IDs
		if html {
			fmt.Println("<ul>")
			for _, t := range cmd.Targets {
				for _, id := range t.Values {
					fmt.Printf("<li><a href='#%s'>%s</a></li>", id, id)
				}
			}
			fmt.Println("</ul>")
		}
		if !quiet {
			fmt.Println(*cmd.Comment)
		}

		// For each target of the command
		for _, target := range cmd.Targets {
			for _, ec2Instance := range target.Values {
				if html {
					fmt.Printf("<h3 id='%s'>%s</h3>", ec2Instance, ec2Instance)
				} else {
					fmt.Println("Getting details for " + ec2Instance)
				}

				// Get command results
				output, err := client.GetCommandInvocation(context.TODO(), &ssm.GetCommandInvocationInput{
					CommandId:  &commandId,
					InstanceId: &ec2Instance,
				})

				if err != nil {
					log.Fatal(err)
				}

				if html {
					fmt.Println(strings.ReplaceAll(*output.StandardOutputContent, "\n", "<br />"))
					fmt.Println("<hr />")
				} else {
					fmt.Println(*output.StandardOutputContent)
				}
			}
		}
	}
}
