# 🍺 Available Beer Visualizer 🍺

ABV is a very simple POS system for beverages with no capacity for handling payment methods. Free as in beer!

## 👮 License

ABV is also *free as in speech* under the [MIT license](LICENSE), so feel free to use it as a baseline for your own projects!

## Usage

### Scanning a new beer

![Scanning New Barcode](https://i.imgur.com/0uDI5yOb.gif)

## 🔧 Configuration

ABV expects a file named config.toml to exist in the ~/.abv directory. An example config file is kept updated in the root of the project repository.

We use [viper](https://github.com/spf13/viper) to handle all configuration. Priority for configuration elements is given to environmental variables, followed by config file entries, followed by default values (which are described when applicable in the example config.toml file)

### 🔒 Untappd Credentials

An ID and Secret are needed to communicate with the Untappd API. These are not included in the config file for obvious security reasons. You can add these to your config file at the root of the toml file named `untappdID` and `untappdSecret` respectively. Alternatively, environmental variables named `ABV_UNTAPPDID` and `ABV_UNTAPPDSECRET` can be used instead.

## 🚀 Deployment

An SQLite database is the heart of the ABV application. The ABV gui can be used to create and update the database. The API application depends on this database but can be run separately as needed. The Frontend application is used to present the HTML5 Menu, and relies on the API to be running.

### 🐳 Docker

Docker containers are uploaded to Docker Hub with the names ``bhutch29/abv_api` and `bhutch29/abv_frontend`. They can be started with the following commands:

`docker run -p 8081:8081 bhutch29/abv_api`

`docker run -p 80:8080 bhutch29/abv_frontend`

Don't forget the `-d` option to start the containers detached from the terminal.

There is also a docker-compose.yml file in the repository root director. The only **prerequisite** for the docker-compose file is that the abv.sqlite database already exist in your ~/.abv directory. This can be accomplished by simply starting the abv application. Then you can launch the above containers using:

`docker-compose up`

Again, don't forget the `-d` option for starting detached!

## 🍻 Contributing

See [contributor guidelines](CONTRIBUTING.md).
