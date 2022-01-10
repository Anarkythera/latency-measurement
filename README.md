# latency-measurement

## Usage

This application requires that the env variable API_KEY exists and is set, this refers to the api key to create an ably client.

```sh
make buildImage
make test
```

The above will run a simple example, where 4 containers will run, each running a different instance of this command line utility.
In the end the results will be saved to a results folder (this folder will be created if it doesn't exist).

## Flags documentation

Currently the latencychecker utility supports the following flags:

- c, to set the name of the channel (default testChannel)
- n, Name to identify the client when sending messages (default $HOSTNAME)
- f, Path where to save the Output (default /tmp)
- m, How many messages to send (default 5)
- w, How long to listen for responses (default 30)
- d, Delay between sending messages in seconds (default 5)
