package cblib

import (
	"fmt"
	"strings"

	cb "github.com/clearblade/Go-SDK"
)

func pullAssets(systemInfo *System_meta, client *cb.DevClient, assets AffectedAssets) (bool, error) {

	didSomething := false

	if assets.UserSchema || assets.AllAssets {
		didSomething = true
		logInfo("Pulling user schema")
		if _, err := pullUserSchemaInfo(systemInfo.Key, client, true); err != nil {
			logError(fmt.Sprintf("Failed to pull user schema - %s\n", err.Error()))
		}
		fmt.Printf("\n")
	}

	if (assets.AllUsers || assets.AllAssets) && assets.ExportUsers {
		didSomething = true
		logInfo("Pulling all users")
		if _, err := PullAndWriteUsers(systemInfo.Key, PULL_ALL_USERS, client, true); err != nil {
			logError(fmt.Sprintf("Failed to pull all users - %s\n", err.Error()))
		}
		if _, err := pullUserSchemaInfo(systemInfo.Key, client, true); err != nil {
			logError(fmt.Sprintf("Failed to pull user schema - %s\n", err.Error()))
		}
		fmt.Printf("\n")
	}

	if assets.User != "" {
		didSomething = true
		logInfo(fmt.Sprintf("Pulling user %+s\n", User))
		_, err := PullAndWriteUsers(systemInfo.Key, User, client, true)
		if err != nil {
			logError(fmt.Sprintf("Failed to pull users. %s", err.Error()))
		}
		if _, err := pullUserSchemaInfo(systemInfo.Key, client, true); err != nil {
			logError(fmt.Sprintf("Failed to pull user schema. %s", err.Error()))
			return false, err
		}
		fmt.Printf("\n")
	}

	if assets.AllServices || assets.AllAssets {
		didSomething = true
		logInfo("Pulling all services")
		if _, err := PullServices(systemInfo.Key, client); err != nil {
			logError(fmt.Sprintf("Failed to pull services. %s", err.Error()))
		}
		fmt.Printf("\n")
	}

	if assets.ServiceName != "" {
		didSomething = true
		logInfo(fmt.Sprintf("Pulling service %+s\n", assets.ServiceName))
		if err := PullAndWriteService(systemInfo.Key, assets.ServiceName, client); err != nil {
			logError(fmt.Sprintf("Failed to pull service. %s", err.Error()))
		}
		fmt.Printf("\n")
	}

	if assets.AllLibraries || assets.AllAssets {
		didSomething = true
		logInfo("Pulling all libraries")
		if _, err := PullLibraries(systemInfo, client); err != nil {
			logError(fmt.Sprintf("Failed to pull libraries. %s", err.Error()))
		}
		fmt.Printf("\n")
	}

	if assets.LibraryName != "" {
		didSomething = true
		logInfo(fmt.Sprintf("Pulling library %s\n", assets.LibraryName))
		if lib, err := pullLibrary(systemInfo.Key, assets.LibraryName, client); err != nil {
			logError(fmt.Sprintf("Failed to pull library. %s", err.Error()))
		} else {
			writeLibrary(lib["name"].(string), lib)
		}
		fmt.Printf("\n")
	}

	if assets.AllCollections || assets.AllAssets {
		didSomething = true
		logInfo("Pulling all collections")
		if _, err := PullAndWriteCollections(systemInfo, client, true, assets.ExportRows, assets.ExportItemId); err != nil {
			logError(fmt.Sprintf("Failed to pull all collections. %s", err.Error()))
		}
		fmt.Printf("\n")
	}

	if assets.CollectionSchema != "" {
		didSomething = true
		logInfo(fmt.Sprintf("Pulling collection schema for %s\n", CollectionSchema))
		if _, err := pullAndWriteCollectionColumns(systemInfo, client, CollectionSchema); err != nil {
			logError(fmt.Sprintf("Failed to pull collection schema. %s", err.Error()))
		}
		fmt.Printf("\n")
	}

	if assets.CollectionName != "" {
		didSomething = true
		logInfo(fmt.Sprintf("Pulling collection %+s\n", CollectionName))
		err := PullAndWriteCollection(systemInfo, CollectionName, client, assets.ExportRows, assets.ExportItemId)
		if err != nil {
			logError(fmt.Sprintf("Failed to pull collection. %s", err.Error()))
		}
		fmt.Printf("\n")
	}

	if assets.AllRoles || assets.AllAssets {
		didSomething = true
		logInfo("Pulling all roles:")
		if _, err := PullAndWriteRoles(systemInfo.Key, client, true); err != nil {
			logError(fmt.Sprintf("Failed to pull all roles. %s", err.Error()))
		}
		fmt.Printf("\n")
	}

	if assets.RoleName != "" {
		didSomething = true
		roles := make([]map[string]interface{}, 0)
		splitRoles := strings.Split(RoleName, ",")
		for _, role := range splitRoles {
			logInfo(fmt.Sprintf("Pulling role %+s\n", role))
			if r, err := pullRole(systemInfo.Key, role, client); err != nil {
				logError(fmt.Sprintf("Failed to pull role. %s", err.Error()))
			} else {
				roles = append(roles, r)
				writeRole(role, r)
			}
		}
		fmt.Printf("\n")
	}

	if assets.AllTriggers || assets.AllAssets {
		didSomething = true
		logInfo("Pulling all triggers")
		if _, err := PullAndWriteTriggers(systemInfo, client); err != nil {
			logError(fmt.Sprintf("Failed to pull all triggers. %s", err.Error()))
		}
		fmt.Printf("\n")
	}

	if assets.TriggerName != "" {
		didSomething = true
		logInfo(fmt.Sprintf("Pulling trigger %+s\n", TriggerName))
		err := PullAndWriteTrigger(systemInfo.Key, TriggerName, client)
		if err != nil {
			logError(fmt.Sprintf("Failed to pull trigger. %s", err.Error()))
		}
		fmt.Printf("\n")
	}

	if assets.AllTimers || assets.AllAssets {
		didSomething = true
		logInfo("Pulling all timers")
		if _, err := PullAndWriteTimers(systemInfo, client); err != nil {
			logError(fmt.Sprintf("Failed to pull all timers. %s", err.Error()))
		}
		fmt.Printf("\n")
	}

	if assets.TimerName != "" {
		didSomething = true
		logInfo(fmt.Sprintf("Pulling timer %+s\n", TimerName))
		err := PullAndWriteTimer(systemInfo.Key, TimerName, client)
		if err != nil {
			logError(fmt.Sprintf("Failed to pull timer. %s", err.Error()))
		}
		fmt.Printf("\n")
	}

	if assets.DeviceSchema || assets.AllAssets {
		didSomething = true
		logInfo("Pulling device schema")
		if _, err := pullDevicesSchema(systemInfo.Key, client, true); err != nil {
			logError(fmt.Sprintf("Failed to pull device schema. %s\n", err.Error()))
		}
		fmt.Printf("\n")
	}

	if assets.AllDevices || assets.AllAssets {
		didSomething = true
		logInfo("Pulling all devices")
		if _, err := PullDevices(systemInfo, client); err != nil {
			logError(fmt.Sprintf("Failed to pull all devices. %s", err.Error()))
		}
		if _, err := pullDevicesSchema(systemInfo.Key, client, true); err != nil {
			logError(fmt.Sprintf("Failed to pull device schema. %s", err.Error()))
		}
		fmt.Printf("\n")
	}

	if assets.DeviceName != "" {
		didSomething = true
		logInfo(fmt.Sprintf("Pulling device %+s\n", DeviceName))
		if device, err := pullDevice(systemInfo.Key, DeviceName, client); err != nil {
			logError(fmt.Sprintf("Failed to pull device. %s", err.Error()))
		} else {
			if _, err := pullDevicesSchema(systemInfo.Key, client, true); err != nil {
				logError(fmt.Sprintf("Failed to pull device schema. %s", err.Error()))
			}
			if err := writeDevice(DeviceName, device); err != nil {
				logError(fmt.Sprintf("Failed to write device. %s", err.Error()))
			}
			roles, err := pullDeviceRoles(systemInfo.Key, DeviceName, client)
			if err != nil {
				logError(fmt.Sprintf("Failed to pull device roles. %s", err.Error()))
			}
			if err := writeDeviceRoles(DeviceName, roles); err != nil {
				logError(fmt.Sprintf("Failed to write device roles. %s", err.Error()))
			}
		}
		fmt.Printf("\n")
	}

	if assets.EdgeSchema || assets.AllAssets {
		didSomething = true
		logInfo("Pulling edge schema")
		if _, err := pullEdgesSchema(systemInfo.Key, client, true); err != nil {
			logError(fmt.Sprintf("Failed to pull edge schema. %s", err.Error()))
		}
		fmt.Printf("\n")
	}

	if assets.AllEdges || assets.AllAssets {
		didSomething = true
		logInfo("Pulling all edges")
		if _, err := PullEdges(systemInfo, client); err != nil {
			logError(fmt.Sprintf("Failed to pull all edges. %s", err.Error()))
		}
		if _, err := pullEdgesSchema(systemInfo.Key, client, true); err != nil {
			logError(fmt.Sprintf("Failed to pull edge schema. %s", err.Error()))
		}
		fmt.Printf("\n")
	}

	if assets.EdgeName != "" {
		didSomething = true
		logInfo(fmt.Sprintf("Pulling edge %+s\n", EdgeName))
		if edge, err := pullEdge(systemInfo.Key, EdgeName, client); err != nil {
			logError(fmt.Sprintf("Failed to pull edge. %s", err.Error()))
		} else {
			writeEdge(EdgeName, edge)
		}
		if _, err := pullEdgesSchema(systemInfo.Key, client, true); err != nil {
			logError(fmt.Sprintf("Failed to pull edge schema. %s", err.Error()))
		}
		fmt.Printf("\n")
	}

	if assets.AllPortals || assets.AllAssets {
		didSomething = true
		logInfo("Pulling all portals")
		if _, err := PullPortals(systemInfo, client); err != nil {
			logError(fmt.Sprintf("Failed to pull all portals. %s", err.Error()))
		}
		fmt.Printf("\n")
	}

	if assets.PortalName != "" {
		didSomething = true
		logInfo(fmt.Sprintf("Pulling portal %+s\n", PortalName))
		if err := PullAndWritePortal(systemInfo.Key, PortalName, client); err != nil {
			logError(fmt.Sprintf("Failed to pull portal. %s", err.Error()))
		}
		fmt.Printf("\n")
	}

	if assets.AllPlugins || assets.AllAssets {
		didSomething = true
		logInfo("Pulling all plugins")
		if _, err := PullPlugins(systemInfo, client); err != nil {
			logError(fmt.Sprintf("Failed to pull all plugins. %s", err.Error()))
		}
		fmt.Printf("\n")
	}

	if assets.PluginName != "" {
		didSomething = true
		logInfo(fmt.Sprintf("Pulling plugin %+s\n", PluginName))
		if err := PullAndWritePlugin(systemInfo.Key, PluginName, client); err != nil {
			logError(fmt.Sprintf("Failed to pull plugin. %s", err.Error()))
		}
		fmt.Printf("\n")
	}

	if assets.AllAdaptors || assets.AllAssets {
		didSomething = true
		logInfo("Pulling all adapters")
		if err := backupAndCleanDirectory(adaptorsDir); err != nil {
			return false, err
		}
		if err := PullAdaptors(systemInfo, client); err != nil {
			if restoreErr := restoreBackupDirectory(adaptorsDir); restoreErr != nil {
				fmt.Printf("Failed to restore backup directory; %s\n", restoreErr.Error())
			}
			logError(fmt.Sprintf("Failed to pull all adapters. %s", err.Error()))
			return false, err
		}
		if err := removeBackupDirectory(adaptorsDir); err != nil {
			fmt.Printf("Warning: Failed to remove backup directory for '%s'", adaptorsDir)
		}
		fmt.Printf("\n")
	}

	if assets.AdaptorName != "" {
		didSomething = true
		logInfo(fmt.Sprintf("Pulling adapter %+s\n", AdaptorName))
		if err := PullAndWriteAdaptor(systemInfo.Key, AdaptorName, client); err != nil {
			logError(fmt.Sprintf("Failed to pull adapter. %s", err.Error()))
		}
		fmt.Printf("\n")
	}

	if assets.AllDeployments || assets.AllAssets {
		didSomething = true
		logInfo("Pulling all deployments")
		if _, err := pullDeployments(systemInfo, client); err != nil {
			logError(fmt.Sprintf("Failed to pull all deployments. %s", err.Error()))
		}
		fmt.Printf("\n")
	}

	if assets.DeploymentName != "" {
		didSomething = true
		logInfo(fmt.Sprintf("Pulling deployment %+s\n", DeploymentName))
		if _, err := pullAndWriteDeployment(systemInfo, client, DeploymentName); err != nil {
			logError(fmt.Sprintf("Failed to pull deployment. %s", err.Error()))
		}
		fmt.Printf("\n")
	}

	if assets.AllServiceCaches || assets.AllAssets {
		didSomething = true
		logInfo("Pulling all service caches")
		if _, err := pullServiceCaches(systemInfo, client); err != nil {
			logError(fmt.Sprintf("Failed to pull all service caches. %s", err.Error()))
		}
		fmt.Printf("\n")
	}

	if assets.ServiceCacheName != "" {
		didSomething = true
		logInfo(fmt.Sprintf("Pulling service cache %+s\n", ServiceCacheName))
		if _, err := pullAndWriteServiceCache(systemInfo, client, ServiceCacheName); err != nil {
			logError(fmt.Sprintf("Failed to pull service cache. %s", err.Error()))
		}
		fmt.Printf("\n")
	}

	if assets.AllWebhooks || assets.AllAssets {
		didSomething = true
		logInfo("Pulling all webhooks")
		if _, err := pullWebhooks(systemInfo, client); err != nil {
			logError(fmt.Sprintf("Failed to pull all webhooks. %s", err.Error()))
		}
		fmt.Printf("\n")
	}

	if assets.WebhookName != "" {
		didSomething = true
		logInfo(fmt.Sprintf("Pulling webhook %+s\n", WebhookName))
		if _, err := pullAndWriteWebhook(systemInfo, client, WebhookName); err != nil {
			logError(fmt.Sprintf("Failed to pull webhook. %s", err.Error()))
		}
		fmt.Printf("\n")
	}

	return didSomething, nil
}
