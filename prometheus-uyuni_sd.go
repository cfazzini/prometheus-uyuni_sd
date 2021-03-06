package main

import (
  "flag"
  "io/ioutil"
  "fmt"
  "os"
  "time"
  "gopkg.in/yaml.v2"
  "regexp"
  "strings"
)

const DEFAULT_CONFIG_FILE = "prometheus-uyuni_sd.yml"

// ------------
//  Structures
// ------------

type Config struct {
  OutputDir        string
  PollingInterval  int
  Host             string
  User             string
  Pass             string
}

// Result structure
type PromScrapeGroup struct {
  Targets         []string
  Labels          map[string]string
}

// ------------------
//  Helper functions
// ------------------

// Error handler
func fatalErrorHandler(e error, msg string) {
  if e != nil {
    fmt.Printf("ERROR: %s\n", e.Error())
    fmt.Printf("ERROR: %s\n", msg)
    os.Exit(1)
  }
}

func getCombinedFormula(combined map[string]exporterConfig, new map[string]exporterConfig) map[string]exporterConfig {
	for k, v := range new {
		if v.Enabled {
			combined[k] = v
		}
	}
	return combined
}

// Generate Scrape targets for uyuni client systems
func writePromConfigForClientSystems(config Config) (error) {
  apiUrl := "http://" + config.Host + "/rpc/api"
  scrapeGroups := []PromScrapeGroup{}
  token, err := Login(apiUrl, config.User, config.Pass)
  if err != nil {
    fmt.Printf("ERROR - Unable to login to SUSE Manager API: %v\n", err)
    return err;
  }
  clientList, err := ListSystems(apiUrl, token)
  if err != nil {
    fmt.Printf("ERROR - Unable to get list of systems: %v\n", err)
    return err;
  }
  if len(clientList) == 0 {
    fmt.Printf("\tFound 0 systems.\n")
  } else {

    for _, client := range clientList {
      custom_values := make(map[string]string)
      custom_labels := make(map[string]string)
      formulas := map[string]exporterConfig{}
      candidateGroups := []groupDetail{}
      details, err := GetSystemDetails(apiUrl, token, client.Id)
      if err != nil {
        fmt.Printf("ERROR - Unable to get system details: %v\n", err)
        return err;
      }

      // Check if system is to be monitored
      for _, v := range details.Entitlements {
        if v == "monitoring_entitled" {
          // 
          // fqdns, err = getSystemNetwork(apiUrl, token, client.Id)

          custom_values, err = getCustomValues(apiUrl, token, client.Id)
					if err != nil {
            fmt.Printf("getCustomValues failed: %v\n", err)
            return err
          }
					// Get list of groups this system is assigned to
					candidateGroups, err = listSystemGroups(apiUrl, token, client.Id)
					if err != nil {
            fmt.Printf("listSystemGroups failed: %v\n", err)
            return err
          }
          
          groups := []string{}
          for _, g := range candidateGroups {
            if g.Subscribed == 1 {
              groupFormulas, err := getGroupFormulaData(apiUrl, token, g.ID, "prometheus-exporters")
              if err != nil {
                fmt.Printf("getGroupFormualData failed: %v\n", err)
                return err
              }
              formulas = getCombinedFormula(formulas, groupFormulas)
              // replace spaces with dashes on all group names
              groups = append(groups, strings.ToLower(strings.ReplaceAll(g.SystemGroupName, " ", "-")))
            }

          }

          
          // Get system formula list
          systemFormulas, _ := getSystemFormulaData(apiUrl, token, client.Id, "prometheus-exporters")
          if err != nil {
            fmt.Printf("getSystemFormulaData failed: %v\n", err)
            return err
          }
          formulas = getCombinedFormula(formulas, systemFormulas)
          // fmt.Printf("%+v\n",  formulas)

          for _,v := range formulas {
            if v.Enabled {
              // only want custom keys that start with "label_"
              re, _ := regexp.Compile("^label_(.*)")
              for k,v := range custom_values {
                if re.MatchString(k) {
                  s := re.ReplaceAllString(k, `$1`)
                  custom_labels[s] = v 
                }
              }     
              scrapeGroups = append (scrapeGroups, PromScrapeGroup{
                Targets: []string{details.Hostname + ":9100"}, Labels: custom_labels,
              })              
            }
          }
          // if (formulas.PostgresExporter.Enabled) {
          //   scrapeGroups = append (scrapeGroups, PromScrapeGroup{
          //     Targets: []string{fqdns.Hostname + ":9187"}, Labels: map[string]string{"role" : "postgres"},
          //   })
          // }
          //fmt.Printf("\tFound system: %s, %v, FQDN: %v Formulas: %+v CustomLabels: %+v \n", details.Hostname, details.Entitlements, fqdns.Hostname, formulas, custom_labels)
        }
      }
     
    }
  }
  Logout(apiUrl, token)
  ymlPromConfig := []byte{}
  if len (scrapeGroups) > 0 {
    ymlPromConfig, _ = yaml.Marshal(scrapeGroups)
  }
  return ioutil.WriteFile(config.OutputDir+"/uyuni-systems.yml", []byte(ymlPromConfig), 0644)
}

// Generate Scrape targets for uyuni server
func writePromConfigForUyuniServer(config Config) (error) {
  promConfig := []PromScrapeGroup{
    PromScrapeGroup{Targets: []string{
      config.Host+":9100", // node_exporeter
      config.Host+":5556", // jmx_exporter tomcat
      config.Host+":5557", // jmx_exporter taskomatic
      config.Host+":9800", // suma exporter
    }},
    PromScrapeGroup{Targets: []string{
      config.Host+":9187",
    }, Labels: map[string]string{"role" : "postgres"}},
  }
  ymlPromConfig, _ := yaml.Marshal(promConfig)
  return ioutil.WriteFile(config.OutputDir+"/suma-server.yml", []byte(ymlPromConfig), 0644)
}

// ------
//  Main
// ------

func main() {
  // Parse command line arguments
  configFile := flag.String("config", DEFAULT_CONFIG_FILE, "Path to config file")
  flag.Parse()
  config := Config{PollingInterval:  120, OutputDir: "/tmp"} // Set defaults

  // Load configuration file
  dat, err := ioutil.ReadFile(*configFile)
  fatalErrorHandler(err, "Unable to read configuration file - please specify the correct location using --config=file.yml")
  err = yaml.Unmarshal([]byte(dat), &config)
  fatalErrorHandler(err, "Unable to parse configuration file")

  // Output some info about supplied config
  fmt.Printf("Using config file: %v\n", *configFile)
  fmt.Printf("\tSUSE Manager API URL: %v\n", config.Host)
  fmt.Printf("\tpolling interval: %d seconds\n", config.PollingInterval)
  fmt.Printf("\toutput dir: %v\n", config.OutputDir)

  // Generate config for SUSE Manager server (self-monitoring)
  writePromConfigForUyuniServer(config)
  // Loop infinitely in case there is a pooling internal, run once otherwise
  for {
    fmt.Printf("Querying SUSE Manager server API...\n")
    startTime := time.Now()
    err := writePromConfigForClientSystems(config)
    duration := time.Since(startTime)
    if err != nil {
      fmt.Printf("ERROR - Unable to generate config for client systems: %v\n", err)
    } else  {
      fmt.Printf("\tQuery took: %s\n", duration)
      fmt.Printf("Prometheus scrape target configuration updated.\n")
    }
    if config.PollingInterval > 0 {
      time.Sleep(time.Duration(config.PollingInterval) * time.Second)
    } else {
      break
    }
  }
}
