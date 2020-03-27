package main

import (
  "github.com/kolo/xmlrpc"
)

type clientRef struct {
  Id                int               `xmlrpc:"id"`
  Name              string            `xmlrpc:"name"`
}

type clientDetail struct {
  Id                int               `xmlrpc:"id"`
  Hostname          string            `xmlrpc:"hostname"`
  Entitlements      []string          `xmlrpc:"addon_entitlements"`
}

type exporterConfig struct {
  Args              string            `xmlrpc:"args"`
  Enabled           bool              `xmlrpc:"enabled"`
}

type formulaData struct {
  NodeExporter      exporterConfig    `xmlrpc:"node_exporter"`
  PostgresExporter  exporterConfig    `xmlrpc:"postgres_exporter"`
}

type networkInfo struct {
  IP                string            `xmlrpc:"ip"`
  IP6               string            `xmlrpc:"ip6"`
  Hostname          string            `xmlrpc:"hostname"`
}

type groupDetail struct {
  ID                int               `xmlrpc:"id"`
  Subscribed        int               `xmlrpc:"subscribed"`
  SystemGroupName   string            `xmlrpc:"system_group_name"`
}

// Attempt to login in SUSE Manager Server and get an auth token
func Login(host string, user string, pass string) (string, error) {
  client, _ := xmlrpc.NewClient(host, nil)
  var result string
  err := client.Call("auth.login", []interface{}{user, pass}, &result)
  return result, err
}

// Logout from SUSE Manager API
func Logout(host string, token string) (error) {
  client, _ := xmlrpc.NewClient(host, nil)
  err := client.Call("auth.logout", token, nil)
  return err
}

// Get client list
func ListSystems(host string, token string) ([]clientRef, error) {
  client, _ := xmlrpc.NewClient(host, nil)
  var result []clientRef
  err := client.Call("system.listSystems", token, &result)
  return result, err
}

// Get client details
func GetSystemDetails(host string, token string, systemId int) (clientDetail, error) {
  client, _ := xmlrpc.NewClient(host, nil)
  var result clientDetail
  err := client.Call("system.getDetails", []interface{}{token, systemId}, &result)
  return result, err
}

// List client FQDNs
func getSystemNetwork(host string, token string, systemId int) (networkInfo, error) {
  client, _ := xmlrpc.NewClient(host, nil)
  var result networkInfo
  err := client.Call("system.getNetwork", []interface{}{token, systemId}, &result)
  return result, err
}



// get system custom key/values
func getCustomValues(host string, token string, systemId int) (map[string]string, error) {
  client, _ := xmlrpc.NewClient(host, nil)
  var result map[string]string
  err := client.Call("system.getCustomValues", []interface{}{token, systemId}, &result)
  return result, err
}

// get groups for a target system
func listSystemGroups(host string, token string, systemId int) ([]groupDetail, error) {
  client, _ := xmlrpc.NewClient(host, nil)
  var result []groupDetail
  err := client.Call("system.listGroups", []interface{}{token, systemId}, &result)
  return result, err
}

// Get formula data for a given system
func getSystemFormulaData(host string, token string, systemId int, formulaName string) (map[string]exporterConfig, error) {
  client, _ := xmlrpc.NewClient(host, nil)
  var result map[string]exporterConfig
  err := client.Call("formula.getSystemFormulaData", []interface{}{token, systemId, formulaName}, &result)
  return result, err
}

// get formula data for a group
func getGroupFormulaData(host string, token string, groupId int, formulaName string) (map[string]exporterConfig, error) {
  client, _ := xmlrpc.NewClient(host, nil)
  var result map[string]exporterConfig
  err := client.Call("formula.getGroupFormulaData", []interface{}{token, groupId, formulaName}, &result)
  return result, err
}