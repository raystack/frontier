# Configuration

| key                       | type              | default                                | description |
| ------------------------- | ----------------- | -------------------------------------- | ----------- |
| endpoint                  | string            | http://localhost:8080                  |             |
| redirectLogin             | string            | http://localhost:3000                  |             |
| redirectSignup            | string            | http://localhost:3000/signup           |             |
| redirectMagicLinkVerify   | string            | http://localhost:3000/magiclink-verify |             |
| callbackUrl               | string            | http://localhost:3000/callback         |             |
| dateFormat                | string            |                                        |             |
| shortDateFormat           | string            |                                        |             |
| theme                     | 'dark' or 'light' | 'light'                                |             |
| billing.successUrl        | string            | http://localhost:3000/success          |             |
| billing.cancelUrl         | string            | http://localhost:3000/cancel           |             |
| billing.cancelAfterTrial  | boolean           | true                                   |             |
| billing.showPerMonthPrice | boolean           | false                                  |             |
| billing.supportEmail      | string            | ''                                     |             |
| billing.hideDecimals      | boolean           | false                                  |             |
| billing.tokenProductId    | string            | ''                                     |             |
