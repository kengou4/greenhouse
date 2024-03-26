/*
 * SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * This file was auto-generated by openapi-typescript.
 * Do not make direct changes to the file.
 */

export interface paths {
  "/Organization": {
    post: {
      responses: {
        /** @description Organization */
        default: {
          content: never
        }
      }
    }
  }
  "/Cluster": {
    post: {
      responses: {
        /** @description Cluster */
        default: {
          content: never
        }
      }
    }
  }
  "/PluginConfig": {
    post: {
      responses: {
        /** @description PluginConfig */
        default: {
          content: never
        }
      }
    }
  }
  "/TeamMembership": {
    post: {
      responses: {
        /** @description TeamMembership */
        default: {
          content: never
        }
      }
    }
  }
  "/Team": {
    post: {
      responses: {
        /** @description Team */
        default: {
          content: never
        }
      }
    }
  }
  "/Plugin": {
    post: {
      responses: {
        /** @description Plugin */
        default: {
          content: never
        }
      }
    }
  }
}

export type webhooks = Record<string, never>

export interface components {
  schemas: {
    /**
     * Organization
     * @description Organization is the Schema for the organizations API
     */
    Organization: {
      /** @description APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
      apiVersion?: string
      /** @description Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
      kind?: string
      metadata?: {
        name?: string
        namespace?: string
        /** Format: uuid */
        uid?: string
        resourceVersion?: string
        /** Format: date-time */
        creationTimestamp?: string
        /** Format: date-time */
        deletionTimestamp?: string
        labels?: {
          [key: string]: string
        }
        annotations?: {
          [key: string]: string
        }
      }
      /** @description OrganizationSpec defines the desired state of Organization */
      spec?: {
        /** @description Authentication configures the organizations authentication mechanism. */
        authentication?: {
          /** @description OIDConfig configures the OIDC provider. */
          oidc?: {
            /** @description ClientIDReference references the Kubernetes secret containing the client id. */
            clientIDReference: {
              /** @description Key in the secret to select the value from. */
              key: string
              /** @description Name of the secret in the same namespace. */
              name: string
            }
            /** @description ClientSecretReference references the Kubernetes secret containing the client secret. */
            clientSecretReference: {
              /** @description Key in the secret to select the value from. */
              key: string
              /** @description Name of the secret in the same namespace. */
              name: string
            }
            /** @description Issuer is the URL of the identity service. */
            issuer: string
            /** @description RedirectURI is the redirect URI. If none is specified, the Greenhouse ID proxy will be used. */
            redirectURI?: string
          }
        }
        /** @description Description provides additional details of the organization. */
        description?: string
        /** @description DisplayName is an optional name for the organization to be displayed in the Greenhouse UI. Defaults to a normalized version of metadata.name. */
        displayName?: string
        /** @description MappedOrgAdminIDPGroup is the IDP group ID identifying org admins */
        mappedOrgAdminIdPGroup?: string
      }
      /** @description OrganizationStatus defines the observed state of an Organization */
      status?: Record<string, never>
    }
    /**
     * Cluster
     * @description Cluster is the Schema for the clusters API
     */
    Cluster: {
      /** @description APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
      apiVersion?: string
      /** @description Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
      kind?: string
      metadata?: {
        name?: string
        namespace?: string
        /** Format: uuid */
        uid?: string
        resourceVersion?: string
        /** Format: date-time */
        creationTimestamp?: string
        /** Format: date-time */
        deletionTimestamp?: string
        labels?: {
          [key: string]: string
        }
        annotations?: {
          [key: string]: string
        }
      }
      /** @description ClusterSpec defines the desired state of the Cluster. */
      spec?: {
        /**
         * @description AccessMode configures how the cluster is accessed from the Greenhouse operator.
         * @enum {string}
         */
        accessMode: "direct" | "headscale"
      }
      /** @description ClusterStatus defines the observed state of Cluster */
      status?: {
        /**
         * Format: date-time
         * @description BearerTokenExpirationTimestamp reflects the expiration timestamp of the bearer token used to access the cluster.
         */
        bearerTokenExpirationTimestamp?: string
        /** @description HeadScaleStatus contains the current status of the headscale client. */
        headScaleStatus?: {
          /** Format: date-time */
          createdAt?: string
          /** Format: date-time */
          expiry?: string
          forcedTags?: string[]
          /** Format: int64 */
          id?: number
          ipAddresses?: string[]
          name?: string
          online?: boolean
          /** @description PreAuthKey reflects the status of the pre-authentication key used by the Headscale machine. */
          preAuthKey?: {
            /** Format: date-time */
            createdAt?: string
            ephemeral?: boolean
            /** Format: date-time */
            expiration?: string
            id?: string
            reusable?: boolean
            used?: boolean
            user?: string
          }
        }
        /** @description KubernetesVersion reflects the detected Kubernetes version of the cluster. */
        kubernetesVersion?: string
        /** @description Nodes provides a map of cluster node names to node statuses */
        nodes?: {
          [key: string]: {
            /** @description Fast track to the node ready condition. */
            ready?: boolean
            /** @description We mirror the node conditions here for faster reference */
            statusConditions?: {
              conditions?: {
                /**
                 * Format: date-time
                 * @description LastTransitionTime is the last time the condition transitioned from one status to another.
                 */
                lastTransitionTime: string
                /** @description Message is a human readable message indicating details about the last transition. May be empty. */
                message?: string
                /** @description Status of the condition. */
                status: string
                /** @description Type of the condition. */
                type: string
              }[]
            }
          }
        }
        /** @description StatusConditions contain the different conditions that constite the status of the cluster. */
        statusConditions?: {
          conditions?: {
            /**
             * Format: date-time
             * @description LastTransitionTime is the last time the condition transitioned from one status to another.
             */
            lastTransitionTime: string
            /** @description Message is a human readable message indicating details about the last transition. May be empty. */
            message?: string
            /** @description Status of the condition. */
            status: string
            /** @description Type of the condition. */
            type: string
          }[]
        }
      }
    }
    /**
     * PluginConfig
     * @description PluginConfig is the Schema for the pluginconfigs API
     */
    PluginConfig: {
      /** @description APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
      apiVersion?: string
      /** @description Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
      kind?: string
      metadata?: {
        name?: string
        namespace?: string
        /** Format: uuid */
        uid?: string
        resourceVersion?: string
        /** Format: date-time */
        creationTimestamp?: string
        /** Format: date-time */
        deletionTimestamp?: string
        labels?: {
          [key: string]: string
        }
        annotations?: {
          [key: string]: string
        }
      }
      /** @description PluginConfigSpec defines the desired state of PluginConfig */
      spec?: {
        /** @description ClusterName is the name of the cluster the pluginConfig is deployed to. If not set, the pluginConfig is deployed to the greenhouse cluster. */
        clusterName?: string
        /** @description Disabled indicates that the plugin config is administratively disabled. */
        disabled: boolean
        /** @description DisplayName is an optional name for the plugin to be displayed in the Greenhouse UI. This is especially helpful to distinguish multiple instances of a Plugin in the same context. Defaults to a normalized version of metadata.name. */
        displayName?: string
        /** @description Values are the values for a plugin instance. */
        optionValues?: {
          /** @description Name of the values. */
          name: string
          /** @description Value is the actual value in plain text. */
          value?: unknown
          /** @description ValueFrom references a potentially confidential value in another source. */
          valueFrom?: {
            /** @description Secret references the secret containing the value. */
            secret?: {
              /** @description Key in the secret to select the value from. */
              key: string
              /** @description Name of the secret in the same namespace. */
              name: string
            }
          }
        }[]
        /** @description Plugin is the name of the plugin this instance is for. */
        plugin: string
      }
      /** @description PluginConfigStatus defines the observed state of PluginConfig */
      status?: {
        /** @description Description provides additional details of the plugin. */
        description?: string
        /** @description ExposedServices provides an overview of the PluginConfigs services that are centrally exposed. It maps the exposed URL to the service found in the manifest. */
        exposedServices?: {
          [key: string]: {
            /** @description Name is the name of the service in the target cluster. */
            name: string
            /** @description Namespace is the namespace of the service in the target cluster. */
            namespace: string
            /**
             * Format: int32
             * @description Port is the port of the service.
             */
            port: number
            /** @description Protocol is the protocol of the service. */
            protocol?: string
          }
        }
        /** @description HelmChart contains a reference the helm chart used for the deployed plugin version. */
        helmChart?: {
          /** @description Name of the HelmChart chart. */
          name: string
          /** @description Repository of the HelmChart chart. */
          repository: string
          /** @description Version of the HelmChart chart. */
          version: string
        }
        /** @description HelmReleaseStatus reflects the status of the latest HelmChart release. This is only configured if the plugin is backed by HelmChart. */
        helmReleaseStatus?: {
          /**
           * Format: date-time
           * @description FirstDeployed is the timestamp of the first deployment of the release.
           */
          firstDeployed?: string
          /**
           * Format: date-time
           * @description LastDeployed is the timestamp of the last deployment of the release.
           */
          lastDeployed?: string
          /** @description Status is the status of a HelmChart release. */
          status: string
        }
        /** @description Message provides a human readbale string abour the current state of the plugin. */
        message?: string
        /**
         * @description Status contains the current state of the plugin config.
         * @enum {string}
         */
        state?: "configured" | "incomplete" | "error"
        /** @description UIApplication contains a reference to the frontend that is used for the deployed plugin version. */
        uiApplication?: {
          /** @description Name of the UI application. */
          name: string
          /** @description URL specifies the url to a built javascript asset. By default, assets are loaded from the Juno asset server using the provided name and version. */
          url?: string
          /** @description Version of the frontend application. */
          version: string
        }
        /** @description Version contains the latest plugin version the config was last applied with successfully. */
        version?: string
        /**
         * Format: int32
         * @description Weight configures the order in which Plugins are shown in the Greenhouse UI.
         */
        weight?: number
      }
    }
    /**
     * TeamMembership
     * @description TeamMembership is the Schema for the teammemberships API
     */
    TeamMembership: {
      /** @description APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
      apiVersion?: string
      /** @description Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
      kind?: string
      metadata?: {
        name?: string
        namespace?: string
        /** Format: uuid */
        uid?: string
        resourceVersion?: string
        /** Format: date-time */
        creationTimestamp?: string
        /** Format: date-time */
        deletionTimestamp?: string
        labels?: {
          [key: string]: string
        }
        annotations?: {
          [key: string]: string
        }
      }
      /** @description TeamMembershipSpec defines the desired state of TeamMembership */
      spec?: {
        /** @description Members list users that are part of a team. */
        members?: {
          /** @description Email of the user. */
          email: string
          /** @description FirstName of the user. */
          firstName: string
          /** @description ID is the unique identifier of the user. */
          id: string
          /** @description LastName of the user. */
          lastName: string
        }[]
      }
      /** @description TeamMembershipStatus defines the observed state of TeamMembership */
      status?: {
        /**
         * Format: date-time
         * @description LastSyncedTime is the information when was the last time the membership was synced
         */
        lastSyncedTime?: string
        /**
         * Format: date-time
         * @description LastChangedTime is the information when was the last time the membership was actually changed
         */
        lastUpdateTime?: string
      }
    }
    /**
     * Team
     * @description Team is the Schema for the teams API
     */
    Team: {
      /** @description APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
      apiVersion?: string
      /** @description Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
      kind?: string
      metadata?: {
        name?: string
        namespace?: string
        /** Format: uuid */
        uid?: string
        resourceVersion?: string
        /** Format: date-time */
        creationTimestamp?: string
        /** Format: date-time */
        deletionTimestamp?: string
        labels?: {
          [key: string]: string
        }
        annotations?: {
          [key: string]: string
        }
      }
      /** @description TeamSpec defines the desired state of Team */
      spec?: {
        /** @description Description provides additional details of the team. */
        description?: string
        /** @description IdP group id matching team. */
        mappedIdPGroup?: string
      }
      /** @description TeamStatus defines the observed state of Team */
      status?: Record<string, never>
    }
    /**
     * Plugin
     * @description Plugin is the Schema for the plugins API
     */
    Plugin: {
      /** @description APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
      apiVersion?: string
      /** @description Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
      kind?: string
      metadata?: {
        name?: string
        namespace?: string
        /** Format: uuid */
        uid?: string
        resourceVersion?: string
        /** Format: date-time */
        creationTimestamp?: string
        /** Format: date-time */
        deletionTimestamp?: string
        labels?: {
          [key: string]: string
        }
        annotations?: {
          [key: string]: string
        }
      }
      /** @description PluginSpec defines the desired state of Plugin */
      spec?: {
        /** @description Description provides additional details of the plugin. */
        description?: string
        /** @description HelmChart specifies where the Helm Chart for this plugin can be found. */
        helmChart?: {
          /** @description Name of the HelmChart chart. */
          name: string
          /** @description Repository of the HelmChart chart. */
          repository: string
          /** @description Version of the HelmChart chart. */
          version: string
        }
        /** @description RequiredValues is a list of values required to create an instance of this Plugin. */
        options?: {
          /** @description Default provides a default value for the option */
          default?: unknown
          /** @description Description provides a human-readable text for the value as shown in the UI. */
          description?: string
          /** @description DisplayName provides a human-readable label for the configuration option */
          displayName?: string
          /** @description Name/Key of the config option. */
          name: string
          /** @description Regex specifies a match rule for validating configuration options. */
          regex?: string
          /** @description Required indicates that this config option is required */
          required: boolean
          /**
           * @description Type of this configuration option.
           * @enum {string}
           */
          type: "string" | "secret" | "bool" | "int" | "list" | "map"
        }[]
        /** @description UIApplication specifies a reference to a UI application */
        uiApplication?: {
          /** @description Name of the UI application. */
          name: string
          /** @description URL specifies the url to a built javascript asset. By default, assets are loaded from the Juno asset server using the provided name and version. */
          url?: string
          /** @description Version of the frontend application. */
          version: string
        }
        /** @description Version of this plugin */
        version: string
        /**
         * Format: int32
         * @description Weight configures the order in which Plugins are shown in the Greenhouse UI. Defaults to alphabetical sorting if not provided or on conflict.
         */
        weight?: number
      }
      /** @description PluginStatus defines the observed state of Plugin */
      status?: Record<string, never>
    }
  }
  responses: never
  parameters: never
  requestBodies: never
  headers: never
  pathItems: never
}

export type $defs = Record<string, never>

export type external = Record<string, never>

export type operations = Record<string, never>