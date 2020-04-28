# Readme
The full source code for the article [Cadence Batcher Example](https://medium.com/stashaway-engineering/). 
Check out the article for a detailed walk-through of this repository. 

In this repo, we are also making use of the cadence adapter that we have wrote previously so check out
[Building your first Cadence Workflow](https://medium.com/stashaway-engineering/building-your-first-cadence-workflow-e61a0b29785) if you havent.

## Running the Server

If you do not have your own cadence docker setup, you can use the one from my cadence 
api example [repository](https://github.com/nndd91/cadence-api-example).

1. `git clone https://github.com/nndd91/cadence-api-example`
2. `cd cadence-api-example/cadenceservice`
3. `docker-compose up`
4. Register the `simple-domain` with `docker run --network=host --rm ubercadence/cli:master --do simple-domain domain register --rd 10`

## Worker and API Server

Navigate back to the project root folder. Make sure go is installed in system.

* Worker
    1. `make worker`
    2. `./bins/worker`

## To start workflow
```
docker run --network=host --rm ubercadence/cli:master --domain simple-domain workflow start --tl "batcherTask" --wt BatcherWorkflow --et 600 --dt 600 -w BatcherMain -i '[]'
```

## To Signal Workflow
```
docker run --network=host --rm ubercadence/cli:master --domain simple-domain wf signal -w BatcherMain -n batcherSignal -i '"customer1"'
```