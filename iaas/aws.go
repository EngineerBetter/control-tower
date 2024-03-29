package iaas

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/route53"
)

// AWSDBSizes maps user set size to RDS instance classes
var AWSDBSizes = map[string]string{
	"small":   "db.t3.small",
	"medium":  "db.t3.medium",
	"large":   "db.m4.large",
	"xlarge":  "db.m4.xlarge",
	"2xlarge": "db.m4.2xlarge",
	"4xlarge": "db.m4.4xlarge",
}

// AWSProvider is the concrete implementation of AWS Provider
type AWSProvider struct {
	sess *session.Session
}

// IEC2 only implements functions used in the iaas package
type IEC2 interface {
	DescribeSecurityGroups(input *ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error)
	DescribeVolumes(input *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error)
	DeleteVolume(input *ec2.DeleteVolumeInput) (*ec2.DeleteVolumeOutput, error)
}

func newAWS(region string) (Provider, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, err
	}
	return &AWSProvider{sess}, nil
}

// DBType gets the correct RDSInstance class
func (a *AWSProvider) DBType(name string) string {
	return AWSDBSizes[name]
}

// Choose for the consumer the appropriate output based on the provider
func (a *AWSProvider) Choose(c Choice) interface{} {
	return c.AWS
}

func (a *AWSProvider) Zone(requestedZone, workerSize string) string {
	if requestedZone != "" {
		return requestedZone
	}
	ec2Client := ec2.New(a.sess)

	zones, err := a.listZones()
	if err != nil {
		return fmt.Sprintf("%sa", a.Region())
	}

	for _, z := range zones {
		o, _ := ec2Client.DescribeReservedInstancesOfferings(&ec2.DescribeReservedInstancesOfferingsInput{
			AvailabilityZone: aws.String(z),
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("instance-type"),
					Values: []*string{aws.String(workerSize)},
				},
			},
		})
		if len(o.ReservedInstancesOfferings) > 0 {
			fmt.Printf("Proposed zone for %s worker instances: %s\n", workerSize, z)
			return z
		}
	}
	return fmt.Sprintf("%sa", a.Region())
}

// Attr returns an attribute of the provider
func (a *AWSProvider) Attr(name string) (string, error) {
	return "", nil
}

// Region returns the region to operate against
func (a *AWSProvider) Region() string {
	return *a.sess.Config.Region
}

// IAAS returns the iaas to operate against
func (a *AWSProvider) IAAS() Name {
	return AWS
}

func (a *AWSProvider) listZones() ([]string, error) {
	ec2Client := ec2.New(a.sess)
	zones := []string{}

	o, err := ec2Client.DescribeAvailabilityZones(&ec2.DescribeAvailabilityZonesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("region-name"),
				Values: []*string{aws.String(a.Region())},
			},
			{
				Name:   aws.String("state"),
				Values: []*string{aws.String("available")},
			},
		},
	})
	if err != nil {
		return zones, err
	}
	for _, z := range o.AvailabilityZones {
		zones = append(zones, *z.ZoneName)
	}
	return zones, nil
}

// CheckForWhitelistedIP checks if the specified IP is whitelisted in the security group
func (a *AWSProvider) CheckForWhitelistedIP(ip, securityGroup string) (bool, error) {

	parsedIP := net.ParseIP(ip)

	ec2Client := ec2.New(a.sess)

	securityGroupsOutput, err := ec2Client.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
		GroupIds: []*string{
			aws.String(securityGroup),
		},
	})
	if err != nil {
		return false, err
	}

	ingressPermissions := securityGroupsOutput.SecurityGroups[0].IpPermissions

	port22, port6868, port25555 := false, false, false
	for _, entry := range ingressPermissions {
		for _, sgIP := range entry.IpRanges {
			_, parsedCIDR, err := net.ParseCIDR(*sgIP.CidrIp)
			if err != nil {
				return false, err
			}
			// support "All traffic rules"
			if *entry.IpProtocol == "-1" {
				if parsedCIDR.Contains(parsedIP) {
					return true, nil
				}
			} else {
				checkPorts(parsedCIDR, parsedIP, &port22, &port6868, &port25555, *entry.FromPort, *entry.ToPort)
			}
		}
	}

	if port22 && port6868 && port25555 {
		return true, nil
	}

	return false, nil
}

func checkPorts(cidr *net.IPNet, ip net.IP, port22, port6868, port25555 *bool, fromPort, toPort int64) {
	if cidr.Contains(ip) {
		// support ranges of ports
		if toPort != fromPort {
			*port22 = between(22, fromPort, toPort)
			*port6868 = between(6868, fromPort, toPort)
			*port25555 = between(25555, fromPort, toPort)
		} else {
			switch fromPort {
			case 22:
				*port22 = true
			case 6868:
				*port6868 = true
			case 25555:
				*port25555 = true
			}
		}
	}
}

func between(value, lower, upper int64) bool {
	return (value <= upper && value >= lower)
}

func checkInUseVolumes(ec2Client *ec2.EC2, volumes []*string) error {
	volumesInUse := 1
	volumesFound := []*ec2.Volume{}
	for i := 0; i < 3 && volumesInUse != 0; i++ {
		volumesOutput, err := ec2Client.DescribeVolumes(&ec2.DescribeVolumesInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("status"),
					Values: []*string{
						aws.String("in-use"),
					},
				},
				{
					Name:   aws.String("volume-id"),
					Values: volumes,
				},
			},
		})

		if err != nil {
			return err
		}
		volumesFound = volumesOutput.Volumes
		volumesInUse = len(volumesFound)
		if volumesInUse > 0 && i < 2 {
			fmt.Printf("There are %v volumes in use. Sleeping for 10 seconds.\n", volumesInUse)
			time.Sleep(10 * time.Second)
		}
	}

	if volumesInUse > 0 {
		output := []string{}
		for _, v := range volumesFound {
			output = append(output, *v.VolumeId)
		}
		fmt.Printf("There are still volumes in use: %+v\n", output)
	}

	return nil
}

// DeleteVolumes deletes the specified EBS volumes
func (a *AWSProvider) DeleteVolumes(volumes []string, deleteVolume func(ec2Client IEC2, volumeID *string) error) error {
	if len(volumes) == 0 {
		return nil
	}

	ec2Client := ec2.New(a.sess)

	var pvolumes []*string
	for i := range volumes {
		pvolumes = append(pvolumes, &volumes[i])
	}
	err := checkInUseVolumes(ec2Client, pvolumes)
	if err != nil {
		return err
	}

	volumesOutput, err := ec2Client.DescribeVolumes(&ec2.DescribeVolumesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("status"),
				Values: []*string{
					aws.String("available"),
					aws.String("creating"),
				},
			},
			{
				Name:   aws.String("volume-id"),
				Values: pvolumes,
			},
		},
	})

	if err != nil {
		return err
	}
	volumesToDelete := volumesOutput.Volumes
	for _, volume := range volumesToDelete {
		volumeID := volume.VolumeId
		err = deleteVolume(ec2Client, volumeID)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteVolume deletes an EBS volume with the given ID
func DeleteVolume(ec2Client IEC2, volumeID *string) error {
	fmt.Printf("Deleting volume: %s\n", *volumeID)
	_, err := ec2Client.DeleteVolume(&ec2.DeleteVolumeInput{
		VolumeId: volumeID,
	})
	return err
}

// DeleteVMsInDeployment is a placeholder for a function used with GCP deployments
func (a *AWSProvider) DeleteVMsInDeployment(zone, project, deployment string) error {
	return nil
}

// DeleteVMsInVPC deletes all the VMs in the given VPC
func (a *AWSProvider) DeleteVMsInVPC(vpcID string) ([]string, error) {

	filterName := "vpc-id"
	ec2Client := ec2.New(a.sess)

	resp, err := ec2Client.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: &filterName,
				Values: []*string{
					&vpcID,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	instancesToTerminate := []*string{}
	volumesToDelete := []string{}
	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			fmt.Printf("Terminating instance %s\n", *instance.InstanceId)
			instancesToTerminate = append(instancesToTerminate, instance.InstanceId)
			for _, blockDevice := range instance.BlockDeviceMappings {
				volumesToDelete = append(volumesToDelete, *blockDevice.Ebs.VolumeId)
			}
		}
	}

	if len(instancesToTerminate) == 0 {
		return nil, nil
	}

	if _, err = ec2Client.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: instancesToTerminate,
	}); err != nil {
		return nil, fmt.Errorf("terminate instances %v: %w", instancesToTerminate, err)
	}

	if err = ec2Client.WaitUntilInstanceTerminatedWithContext(
		context.Background(),
		&ec2.DescribeInstancesInput{InstanceIds: instancesToTerminate},
		func(w *request.Waiter) {
			// Wait 20 minutes, checking every 30 seconds
			w.MaxAttempts = 40
			w.Delay = func(_ int) time.Duration { return time.Second * 30 }
		},
	); err != nil {
		return nil, fmt.Errorf("wait for instance termination: %w", err)
	}

	return volumesToDelete, nil
}

// ListHostedZones returns a list of hosted zones
func (a *AWSProvider) ListHostedZones() ([]*route53.HostedZone, error) {

	r53Client := route53.New(a.sess)
	hostedZones := []*route53.HostedZone{}
	err := r53Client.ListHostedZonesPages(&route53.ListHostedZonesInput{}, func(output *route53.ListHostedZonesOutput, _ bool) bool {
		hostedZones = append(hostedZones, output.HostedZones...)
		return true
	})
	if err != nil {
		return nil, err
	}

	return hostedZones, nil
}

// FindLongestMatchingHostedZone finds the longest hosted zone that matches the given subdomain
func (a *AWSProvider) FindLongestMatchingHostedZone(subdomain string) (string, string, error) {
	hostedZones, err := a.ListHostedZones()
	if err != nil {
		return "", "", err
	}

	longestMatchingHostedZoneName := ""
	longestMatchingHostedZoneID := ""
	for _, hostedZone := range hostedZones {
		domain := strings.TrimRight(*hostedZone.Name, ".")
		id := *hostedZone.Id
		if strings.HasSuffix(subdomain, domain) {
			if len(domain) > len(longestMatchingHostedZoneName) {
				longestMatchingHostedZoneName = domain
				longestMatchingHostedZoneID = id
			}
		}
	}

	if longestMatchingHostedZoneName == "" {
		return "", "", fmt.Errorf("no matching hosted zone found for domain %s", subdomain)
	}

	longestMatchingHostedZoneID = strings.Replace(longestMatchingHostedZoneID, "/hostedzone/", "", -1)

	return longestMatchingHostedZoneName, longestMatchingHostedZoneID, err
}

// CreateDatabases creates databases on the server
func (a *AWSProvider) CreateDatabases(name, username, password string) error {
	return fmt.Errorf("not implemented")
}
