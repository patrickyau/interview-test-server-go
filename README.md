# Interview Test Server

It project demonstrates how the Interview Test Server can calculate tax return. This project has been updated to use the [oapi-codegen tool](https://github.com/oapi-codegen/oapi-codegen) to generate interface code from OpenAPI.yaml file. Implementation code has been added to get all APIs to work.

## The Assignment - Tax Calculator

Your task is to build an application that queries our dockerized API and displays the calculated total income tax for a given salary and tax year.
You may refer to this resource for context [on how to calculate total income tax](https://investinganswers.com/dictionary/m/marginal-tax-rate#:~:text=To%20calculate%20marginal%20tax%20rate) using marginal tax rates.

Please [see the official reference to tax brackets](https://www.canada.ca/en/financial-consumer-agency/services/financial-toolkit/taxes/taxes-2/5.html) and rates for more information.

**Completed**

## What has been done

The goal of this assessment is to provide a picture of your approach to development as it relates to:

* Design patterns/programming paradigm - Would like to introduce Model-View-Controller (MVC - Command) pattern but, ran out of time.
* Scalability
* API interface design
* Frameworks - Gin server framework
* Documentation - Readme and function comments
* Clean code
* UI
* Testing - added some but, not complete
* Automated testing - Added some
* Error handling - completed
* Logging
* Readability

### For backend candidates

The application you’re building should have an HTTP API with an endpoint that takes an annual income and the tax year as parameters. The appropriate type of params (query vs body param vs URL etc.) is to be determined by you. Your endpoint should return a JSON object with the result of the calculation.

**A POST request was introduced to calculate the tax owed. The body of the request can be further expanded to support other information to be submitted to adjust tax amount.
 

### In both cases, it should:

#### Accomplish the following:

* Fetch the tax rates by year i.e. 
  [/tax-calculator/tax-years/[2019|2020|2021|2022]](http://localhost:8080/tax-calculator/tax-years/2022)
* Receive a yearly salary
* Calculate and display the total taxes owed for the salary
* Display the amount of taxes owed per band
* Display the effective rate

**Done**

#### Apply to the following 2022 tax year scenarios:

| Salary      | Total Taxes |
|-------------|-------------|
| $0 <=       | $0          |
| $50,000     | $7,500.00   |
| $100,000    | $17,739.17  |
| $1,234,567  | $385,587.65 |

**Verified manually**

### Sample Request

The sample POST request to `http://localhost:8080/tax-calculator/tax-years/2023/calculate` can use the following JSON as the body

```json
{
  "salary": 1234567
}
```

The smaple response can be found here:
```json
{
    "effective_tax_rate": "31.12%",
    "salary": 1234567,
    "tax_owned_per_band": [
        {
            "min": 0,
            "max": 53359,
            "rate": 0.15,
            "tax_owed": 8003.85
        },
        {
            "min": 53359,
            "max": 106717,
            "rate": 0.205,
            "tax_owed": 10938.39
        },
        {
            "min": 106717,
            "max": 165430,
            "rate": 0.26,
            "tax_owed": 15265.38
        },
        {
            "min": 165430,
            "max": 235675,
            "rate": 0.29,
            "tax_owed": 20371.05
        },
        {
            "min": 235675,
            "max": 0,
            "rate": 0.33,
            "tax_owed": 329634.37
        }
    ],
    "tax_year": "2023",
    "total_tax_owed": 384213.03
}
```


## Get up and running
To build the docker image, please follow these instructions:
```bash
docker build --tag interview-test-server-go .
```
In order to run the API locally, please follow these instructions:

```bash
docker pull patrickyau/interview-test-server-go
docker run --init --rm -p 8080:8080 --name interview-test-server interview-test-server (or use `make run`)
```
OR simply use the `make` command to run the service:
```bash
make run
```

Navigate to [http://localhost:8080/tax-calculator/health](http://localhost:8080/tax-calculator/health). You should be greeted with this message:
```json
{
    "status": "ok"
}
```

To access to the different available endpoints. The following are the relevant endpoints:

* GET [/tax-calculator/](http://localhost:8080/tax-calculator/) - endpoint to get the default 2022 tax rates
* GET [/tax-calculator/tax-years/2022](http://localhost:8080/tax-calculator/tax-years/2022) - endpoint to get the tax rates
* GET [/tax-calculator/tax-years](http://localhost:8080/tax-calculator/tax-years) - endpoint to get the tax rates of all years
* POST [/tax-calculator/tax-years/2022/calculate](http://localhost:8080/tax-calculator/tax-years/2022/calculate) - endpoint to get the tax owed for the year
* GET [/tax-calculator/health](http://localhost:8080/tax-calculator/health) - endpoint to get the health of the service


