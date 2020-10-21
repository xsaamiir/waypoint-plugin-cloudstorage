# Waypoint Plugin Google Cloud Storage [**WIP**]

waypoint-plugin-cloudstorage is a registry plugin for [Waypoint](https://github.com/hashicorp/waypoint). 
It allows you to upload an artifact to Google Cloud Storage.

**This plugin is still work in progress, please open an issue for any feedback or issues.**

# Install
To install the plugin, run the following command:

````bash
git clone git@github.com:sharkyze/waypoint-plugin-cloudstorage.git # or gh repo clone sharkyze/waypoint-plugin-cloudstorage
cd waypoint-plugin-cloudstorage
make install # Installs the plugin in `${HOME}/.config/waypoint/plugins/`
````

# Google Cloud Storage Authentication
Please follow the instructions in the [Google Cloud Run tutorial](https://learn.hashicorp.com/tutorials/waypoint/google-cloud-run?in=waypoint/deploy-google-cloud#authenticate-to-google-cloud).
This plugin uses GCP Application Default Credentials (ADC) for authentication. More info [here](https://cloud.google.com/docs/authentication/production).

# Configure
```hcl
project = "project"

app "webapp" {
  path = "./webapp"

  build {
    use "archive" {
      sources = ["src/", "public/", "package.json"] # Sources are relative to /path/to/project/webapp/
      output_name = "webapp.zip"
      overwrite_existing = true
    }

    registry {
       use "cloudstorage" {
         source = "webapp.zip"
         name = "artifcats/webapp/${gitrefpretty()}.zip"
         bucket = "staging.gcp-project-name.appspot.com"
       }
     }
  }
}
```

