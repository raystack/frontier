# SMTP Server Configurations

Shield can be used to send invites to users to join an organization currently or send OTPs (One Time Password) for verification. For this it implements a mailer service which provides the functionality to send emails using the configured [SMTP(Simple Mail Transfer Protocol)](https://datatracker.ietf.org/doc/html/rfc2821) server. This involves establishing a connection with the SMTP server, authenticating with the provided credentials, and delivering the email to the specified recipients.

This document provides instructions on how to configure the SMTP settings for sending emails using the mailer configuration in Shield.

### Prerequisites
- SMTP server credentials (username and password)
- SMTP server hostname and port information
- Sender's Email address for the "from" header

### Configuration

To configure the SMTP settings for sending emails, follow the steps below:

1. Open the shield server configuration file. Locate the mailer section.
2. Update the following parameters within the mailer section:
    - **smtp_host**: Set this value to the hostname of your SMTP server. For example: smtp.example.com
    - **smtp_port**: Set this value to the port number used by your SMTP server. Typically, this is **587** for secure connections or **25** for insecure connections.
    - **smtp_username**: Enter the username or email address associated with your SMTP server account.
    - **smtp_password**: Enter the password for your SMTP server account.
    - **smtp_insecure**: If your SMTP server does not use a secure connection, set this value to true. If a secure connection (TLS/SSL) is required, set it to false.
    - **headers.from**: Set this value to the email address you want to appear in the "from" header when sending emails. For example, username@acme.org.

Save the configuration file with the updated SMTP settings. And restart the Shield server.

### SMTP Providers 

- **Google Workspace SMTP:** to use Google Workspace SMTP for sending mails, read this [guide](https://support.google.com/a/answer/176600?hl=en)
- **Amazon SES service:** to use Amazon SES for sending mails, follow this [guide](https://docs.aws.amazon.com/ses/latest/dg/send-email-smtp.html) 
- **MailGun:** to use Mailgun, follow this [guide](https://www.miniorange.com/iam/content-library/smtp-gateways/setup-mailgun-as-smtp)
- **PostMark:** to use PostMark SMTP service for sending mails from Shield, read this [guide](https://postmarkapp.com/developer/user-guide/send-email-with-smtp)