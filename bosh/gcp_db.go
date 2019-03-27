package bosh

func (client *GCPClient) createDefaultDatabases() error {
	return client.provider.CreateDatabases(client.config.GetRDSDefaultDatabaseName(), client.config.GetRDSUsername(), client.config.GetRDSPassword())
}
