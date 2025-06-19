package serverless

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pulumi/pulumi-azure-native-sdk/app/v3"
	// "github.com/pulumi/pulumi-azure-native-sdk/authorization/v3"
	"github.com/pulumi/pulumi-azure-native-sdk/managedidentity/v3"
	"github.com/pulumi/pulumi-azure-native-sdk/operationalinsights/v3"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v3"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

func CreateScheduledJob(ctx *pulumi.Context, resourceGroup *resources.ResourceGroup,
	prefix, componentID string,
	cmd string, delay string) error {

	if err := manager.CheckBackedURLForServerless(); err != nil {
		return err
	}

	cronExp := getCronExpressionForScheduledTrigger(delay)

	r := &serverlessRequestArgs{
		prefix:             prefix,
		componentID:        componentID,
		scheduleExpression: cronExp,
		command:            cmd,
		resourceGroup:      resourceGroup,
	}
	return r.deploy(ctx)
}

func (a *serverlessRequestArgs) deploy(ctx *pulumi.Context) error {
	// create userassigned identity to get access to az resources from the container
	uaid, err := managedidentity.NewUserAssignedIdentity(ctx, resourcesUtil.GetResourceName(a.prefix, a.componentID, "uaid"),
		&managedidentity.UserAssignedIdentityArgs{
			Location:          a.resourceGroup.Location,
			ResourceGroupName: a.resourceGroup.Name,
			ResourceName:      pulumi.String(resourcesUtil.GetResourceName(a.prefix, a.componentID, "uaid")),
			Tags:              maptContext.ResourceTags(),
		})
	if err != nil {
		return err
	}

	// create analytics workspace for containerapp logs
	analyticsWorkspace, err := operationalinsights.NewWorkspace(ctx, resourcesUtil.GetResourceName(a.prefix, a.componentID, "log-analytics-workspace"),
		&operationalinsights.WorkspaceArgs{
			Location:          a.resourceGroup.Location,
			ResourceGroupName: a.resourceGroup.Name,
			RetentionInDays:   pulumi.Int(30),
			Sku: &operationalinsights.WorkspaceSkuArgs{
				Name: pulumi.String(operationalinsights.WorkspaceSkuNameEnumPerGB2018),
			},
		})
	if err != nil {
		return err
	}

	// use pulumi all to wait for both resource group and analyticsWorkspace
	// are created, then fetch the shared key

	shk := pulumi.All(analyticsWorkspace.Name, a.resourceGroup.Name).ApplyT(
		func(args []interface{}) (*string, error) {
			// get analytics workspace shared key
			shk, err := operationalinsights.GetWorkspaceSharedKeys(ctx, &operationalinsights.GetWorkspaceSharedKeysArgs{
				WorkspaceName:     args[0].(string),
				ResourceGroupName: args[1].(string),
			})
			if err != nil {
				return nil, err
			}
			return shk.PrimarySharedKey, nil
		}).(pulumi.StringPtrOutput)

	// create containerapp environment
	environment, err := app.NewManagedEnvironment(ctx, resourcesUtil.GetResourceName(a.prefix, a.componentID, "containerapp-environment"),
		&app.ManagedEnvironmentArgs{
			Location:          a.resourceGroup.Location,
			ResourceGroupName: a.resourceGroup.Name,
			AppLogsConfiguration: &app.AppLogsConfigurationArgs{
				Destination: pulumi.String("log-analytics"),
				LogAnalyticsConfiguration: &app.LogAnalyticsConfigurationArgs{
					CustomerId: analyticsWorkspace.CustomerId,
					SharedKey:  shk,
				},
			},
		}, pulumi.DependsOn([]pulumi.Resource{analyticsWorkspace}))
	if err != nil {
		return err
	}

	lcpu, err := strconv.Atoi(LimitCPU)
	if err != nil {
		return err
	}

	principalID := uaid.ID().ApplyT(
		func(id string) string {
			return id
		},
	).(pulumi.StringOutput)

	// add 'Contributor' role to userassigned id to provide permissions to access the resource group
	// _, err = authorization.NewRoleAssignment(ctx, resourcesUtil.GetResourceName(a.prefix, a.componentID, "role-assignment"),
	// 	&authorization.RoleAssignmentArgs{
	// 		PrincipalId:      principalID,
	// 		PrincipalType:    pulumi.String(authorization.PrincipalTypeServicePrincipal),
	// 		RoleDefinitionId: pulumi.String("/providers/Microsoft.Authorization/roleDefinitions/b24988ac-6180-42a0-ab88-20f7382dd24c"),
	// 		Scope:            a.resourceGroup.ID(),
	// 	})
	// if err != nil {
	// 	return err
	// }

	_, err = app.NewJob(ctx, resourcesUtil.GetResourceName(a.prefix, a.componentID, "job"),
		&app.JobArgs{
			Location:          a.resourceGroup.Location,
			ResourceGroupName: a.resourceGroup.Name,
			JobName:           pulumi.String(resourcesUtil.GetResourceName(a.prefix, a.componentID, "destroy-job")),
			EnvironmentId:     environment.ID(),
			Configuration: app.JobConfigurationArgs{
				ScheduleTriggerConfig: &app.JobConfigurationScheduleTriggerConfigArgs{
					CronExpression: pulumi.String(a.scheduleExpression),
				},
				ReplicaRetryLimit: pulumi.Int(10),
				ReplicaTimeout:    pulumi.Int(10),
				TriggerType:       pulumi.String(app.TriggerTypeSchedule),
			},
			Template: &app.JobTemplateArgs{
				Containers: app.ContainerArray{
					&app.ContainerArgs{
						Name: pulumi.String(resourcesUtil.GetResourceName(a.prefix, a.componentID, "mapt-container")),
						Resources: &app.ContainerResourcesArgs{
							Cpu:    pulumi.Float64(lcpu),
							Memory: pulumi.String(LimitMemory),
						},
						Image: pulumi.String(maptContext.OCI),
						Args:  pulumi.ToStringArray(strings.Fields(a.command)),
					},
				},
			},
			Identity: app.ManagedServiceIdentityArgs{
				Type: app.ManagedServiceIdentityTypeUserAssigned,
				UserAssignedIdentities: pulumi.StringArray{
					principalID,
				},
			},
		},
		pulumi.DependsOn([]pulumi.Resource{uaid, environment}),
	)

	if err != nil {
		return err
	}

	ctx.Export("job-name", pulumi.String(resourcesUtil.GetResourceName(a.prefix, a.componentID, "destroy-job")))
	ctx.Export("resource-group-name", a.resourceGroup.Name)
	ctx.Export("container-name", pulumi.String(resourcesUtil.GetResourceName(a.prefix, a.componentID, "mapt-container")))
	return nil
}

func getCronExpressionForScheduledTrigger(timeout string) string {
	duration, err := time.ParseDuration(timeout)
	if err != nil {
		return ""
	}
	tfuture := time.Now().UTC().Add(duration)

	return fmt.Sprintf("%d %d * * *", tfuture.Minute(), tfuture.Hour())
}
