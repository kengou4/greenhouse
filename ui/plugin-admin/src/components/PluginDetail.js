/*
 * SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useState } from "react"
import {
  useShowDetailsFor,
  usePluginActions,
  usePluginConfig,
} from "./StoreProvider"
import {
  CodeBlock,
  Container,
  DataGrid,
  DataGridRow,
  DataGridCell,
  DataGridHeadCell,
  JsonViewer,
  Panel,
  Pill,
  PanelBody,
  Stack,
  Tabs,
  TabList,
  Tab,
  TabPanel,
  Icon,
} from "juno-ui-components"
import { PluginConditionIcon } from "./PluginConditionIcon"

// Renders the plugin details panel
const PluginDetail = () => {
  const pluginConfig = usePluginConfig()
  const { setShowDetailsFor } = usePluginActions()
  const showDetailsFor = useShowDetailsFor()
  const [plugin, setPlugin] = useState(null)

  useEffect(() => {
    if (!showDetailsFor || !pluginConfig) {
      return
    }
    setPlugin(pluginConfig.find((p) => p.id === showDetailsFor))
  }, [showDetailsFor, pluginConfig])

  const onPanelClose = () => {
    setShowDetailsFor(null)
  }

  return (
    <Panel
      opened={!!showDetailsFor}
      onClose={onPanelClose}
      size="large"
      heading={
        <Stack gap="2">
          <PluginConditionIcon plugin={plugin} />
          <span>{plugin?.name}</span>
        </Stack>
      }
    >
      <PanelBody>
        <Tabs>
          <TabList>
            <Tab>Details</Tab>
            <Tab>Raw Data</Tab>
          </TabList>
          <TabPanel>
            <Container px={false} py>
              <DataGrid
                columns={2}
                cellVerticalAlignment="top"
                gridColumnTemplate="35% auto"
              >
                <DataGridRow>
                  <DataGridHeadCell>Name</DataGridHeadCell>
                  <DataGridCell>{plugin?.name}</DataGridCell>
                </DataGridRow>

                {plugin?.disabled && (
                  <DataGridRow>
                    <DataGridHeadCell>Disabled</DataGridHeadCell>
                    <DataGridCell>{plugin?.disabled.toString()}</DataGridCell>
                  </DataGridRow>
                )}

                <DataGridRow>
                  <DataGridHeadCell>Version</DataGridHeadCell>
                  <DataGridCell>{plugin?.version}</DataGridCell>
                </DataGridRow>

                <DataGridRow>
                  <DataGridHeadCell>Cluster</DataGridHeadCell>
                  <DataGridCell>{plugin?.clusterName}</DataGridCell>
                </DataGridRow>

                <DataGridRow>
                  <DataGridHeadCell>External Links</DataGridHeadCell>
                  <DataGridCell>
                    {plugin?.externalServicesUrls?.map((url) => {
                      return (
                        <a
                          href={url.url}
                          target="_blank"
                          rel="noreferrer"
                          className="mr-3"
                          key={url.url}
                        >
                          {url.name}
                        </a>
                      )
                    })}
                  </DataGridCell>
                </DataGridRow>

                <DataGridRow>
                  <DataGridHeadCell>Condition</DataGridHeadCell>
                  <DataGridCell>
                    <PluginConditionIcon plugin={plugin} />
                  </DataGridCell>
                </DataGridRow>

                <DataGridRow>
                  <DataGridHeadCell>Conditions</DataGridHeadCell>
                  <DataGridCell>
                    <Stack gap="2" alignment="start" wrap={true}>
                      {plugin?.statusConditions?.map((condition) => {
                        return (
                          <Pill
                            key={condition.type}
                            pillKey={condition.type}
                            pillValue={condition.status}
                          />
                        )
                      })}
                    </Stack>
                  </DataGridCell>
                </DataGridRow>
                {plugin?.optionValues?.map((option) => {
                  {
                    /* Every optionValue which not starts with greenhouse is shown */
                  }
                  if (option?.name.startsWith("greenhouse.")) return null

                  return (
                    <DataGridRow key={option?.name}>
                      <DataGridHeadCell style={{ overflowWrap: "break-word" }}>
                        {option?.name}
                      </DataGridHeadCell>
                      <DataGridCell>
                        {typeof option?.value != "undefined" &&
                          (typeof option?.value === "object" ? (
                            <JsonViewer data={option?.value} />
                          ) : (
                            String(option?.value)
                          ))}
                      </DataGridCell>
                    </DataGridRow>
                  )
                })}
              </DataGrid>
            </Container>
          </TabPanel>

          <TabPanel>
            <Container px={false} py>
              <CodeBlock>
                <JsonViewer data={plugin?.raw} expanded={true} />
              </CodeBlock>
            </Container>
          </TabPanel>
        </Tabs>
      </PanelBody>
    </Panel>
  )
}

export default PluginDetail
