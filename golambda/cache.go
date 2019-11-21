package golambda

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"net/http"
	"os"
	"strings"
	"sync"
)

type MyEvent struct {
	PostId string `json:"object"`
}

type ResponseItem struct {
	PostId         string `json:"postId"`
	ResponseStatus int    `json:"reposeStatus"`
	Ip             string `json:"ip"`
}

type Response struct {
	Items []ResponseItem
}

func HandleRequest(name MyEvent) (Response, error) {
	if name.PostId == "" {
		panic("post id is empty")
	}
	var items []ResponseItem
	var autoScalingGroupNames []*string
	client := &http.Client{}
	for _, value := range strings.Split(os.Getenv("ASG_LIST"), ", ") {
		autoScalingGroupNames = append(autoScalingGroupNames, aws.String(value))
	}
	instances := getInstancesPublicIps(os.Getenv("REGION"), autoScalingGroupNames)
	wg := new(sync.WaitGroup)

	for _, instance := range instances {
		wg.Add(1)
		go func(ip string, group *sync.WaitGroup) {
			defer group.Done()
			request, err := http.NewRequest("DELETE", fmt.Sprintf("http://%s:80/.*%s.*", ip, name.PostId), nil)
			if err != nil {
				panic(err)
			}
			resp, err := client.Do(request)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			items = append(items, ResponseItem{PostId: name.PostId, ResponseStatus: resp.StatusCode, Ip: ip})
		}(instance, wg)

	}
	wg.Wait()

	return Response{Items: items}, nil
}

func getInstancesPublicIps(awsRegion string, autoScalingGroupNames []*string) []string {
	var instanceIds []*string
	var instanceIps []string
	mySession := session.Must(session.NewSession())

	autoScaling := autoscaling.New(mySession, aws.NewConfig().WithRegion(awsRegion))
	ec2Client := ec2.New(mySession, aws.NewConfig().WithRegion(awsRegion))
	result, err := autoScaling.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{AutoScalingGroupNames: autoScalingGroupNames})
	if err != nil {
		panic(err)
	}
	for _, autoScalingGroup := range result.AutoScalingGroups {
		for _, instance := range autoScalingGroup.Instances {
			instanceIds = append(instanceIds, instance.InstanceId)
		}
	}

	instances, err := ec2Client.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: instanceIds,
	})
	if err != nil {
		panic(err)
	}
MainLoop:
	for {
		for _, reservation := range instances.Reservations {
			for _, instanceReservation := range reservation.Instances {
				instanceIps = append(instanceIps, *instanceReservation.PrivateIpAddress)

			}
		}
		if instances.NextToken == nil {
			break MainLoop
		}
		instances, err = ec2Client.DescribeInstances(&ec2.DescribeInstancesInput{
			InstanceIds: instanceIds,
			NextToken:   instances.NextToken,
		})
		if err != nil {
			panic(err)
		}
	}
	return instanceIps
}
