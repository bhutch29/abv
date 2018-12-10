# Available Beer Visualizer

ABV is a very simple POS system for beverages with no capacity for handling payment methods. Free as in beer! Its also free as in speech under the MIT license, so feel free to use it as a baseline for your own projects!

## Configuration

ABV expects a file named config.toml to exist in the ~/.abv directory. An example config file is kept updated in the root of the project repository.

We use [viper](https://github.com/spf13/viper) to handle all configuration. Priority for configuration elements is given to environmental variables, followed by config file entries, followed by default values (which are described when applicable in the example config.toml file)

### Untappd Credentials

An ID and Secret are needed to communicate with the Untappd API. These are not included in the config file for obvious security reasons. You can add these to your config file at the root of the toml file named `untappdID` and `untappdSecret` respectively. Alternatively, environmental variables named `ABV_UNTAPPDID` and `ABV_UNTAPPDSECRET` can be used instead.

## Deployment

An SQLite database is the heart of the ABV application. The ABV gui can be used to create and update the database. The API application depends on this database but can be run separately as needed. The Frontend application is used to present the HTML5 Menu, and depends on the API to be running.
