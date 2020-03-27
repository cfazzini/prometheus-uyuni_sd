# Prometheus SUSE Manager/Uyuni Service Discovery

This tool connects to SUSE Manager servers and generates Prometheus scrape target configurations, taking advantage of the [file-based service discovery](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#%3Cfile_sd_config) mechanism provided by Prometheus.

## Prometheus Configuration

Please remember to adjust your `prometheus.yml` configuration file to use the file service discovery mechanism and point it to the output location of this tool.

Example configuration section of prometheus.yml:
```yaml
- job_name: 'overwritten-default'
  file_sd_configs:
   - files: ['/data/prometheus/scrape-config/*.yml']
```
## Notes
Added functionality to generate Prometheus target labels from the [Custom System Info](https://www.uyuni-project.org/uyuni-docs/uyuni/reference/systems/custom-system-info.html) feature. Currently hardcoded to scrape keys prefixed with "label_".

## Credits
Forked from: [github.com/cavalheiro/prometheus-suma_sd](https://github.com/cavalheiro/prometheus-suma_sd)

Merged code from this [SUSE Prometheus Patch](https://github.com/bmwiedemann/openSUSE/blob/2646c9ed431d58c9332af97c9283d4bdf304d705/packages/g/golang-github-prometheus-prometheus/0003-Add-Uyuni-service-discovery.patch)

Thanks SUSE for the heavy lifting!

## License

This project has been released under the MIT license. Please see the LICENSE.md file for more details.
