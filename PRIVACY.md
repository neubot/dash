# Privacy Policy

## History

- v3.0.0: current version (this document)
- [v2.0.3](https://github.com/neubot/neubot/blob/master/PRIVACY): previous
  version that predates GDPR and where Nexa was data controller

## Introduction

Neubot is an Internet measurement experiment that focuses on
studying Internet performance since 2010. Users will run Neubot
experiments through applications that integrate a compatible
implementation of the Neubot measurement protocol.

Neubot used to be run by the [Nexa Center for Internet & Society](
https://nexa.polito.it/) and is now run by Simone Basso, an
independent Internet researcher who originally developed Neubot.

Neubot has signed a Memorandum of Understanding (MoU) with [Measurement
Lab](https://www.measurementlab.net/) that allows Neubot to host
measurement server-side software on the distributed fleet of measurement
servers maintaned by Measurement Lab (henceforth M-Lab).

The specification of Neubot experiments, and their reference
implementation, use M-Lab services to determine the M-Lab server
to perform a network experiment with, then runs a network experiment
with such server. Neubot's server-side code saves the results of
such network experiment on the same server. This data is made
available as open data, allowing reuse for any purpose. Code written
by M-Lab and running on the same server, will then harvest the
results, delete the originals, and publish the results as open data.

The results of Neubot experiments may contain personal data. Neubot
is data controller (GDPR Art. 4.7) insofar as it determines the means
and purposes of the data collected and made public.

Because of the aforementioned MoU with M-Lab, and because of the way in
which data is harvested by M-Lab, Neubot shares the control of such data
with M-Lab. Therefore, you are strongly advised to also read the [M-Lab
privacy policy](https://www.measurementlab.net/privacy/).

## What data is collected and why

Data collected by Neubot include network performance metrics, the date
and time when a measurement was performed, the IP address, and possibly
metadata regarding the browser and/or the operating system.

Both the IP address and the date and time of the measurement are required to
properly contextualize the measurement.

Neubot was developed to promote scientific, transparent, nonpartisan analysis
of Internet performance. To achieve this, we believe that scientific
independence requires the resulting data to be made available to the wider
public as raw data without any limitation of time.

This entails storing data forever.

Because Neubot does not have the resources to carry on this mission alone, we
partnered with M-Lab since 2012 to help with server provisioning, data
harvesting and publish.

As a consequence we collect and (facilitate the) publish(ing of) your data
under legitimate interest.

## Your rights under GDPR

As Neubot users, under the GDPR, considering that your personal
data was collected, processed, and published under the legal base
of legitimate interest, you have the following rights:

1. to access your data: you may request that we provide you with
   a copy of the data we hold about you.

2. to request the rectification of your personal data.

3. to request the erasure or restriction of your personal
   data and to object to its processing.

4. to object to automated decision-making and the right to be informed
   about whether or not your data is subject to automated decision making.

Automated decision-making is involved in determining which server
to run a network measurement with, as described above. This
decision-making is in no way employed to make decisions which would
affect your rights and freedoms.

If you contact us by email, you should expect your data to be sent
by email. If you wish to increase your privacy, please secure the
email exchange using PGP. If you would like it to be provided through
another medium, please let us know.

Neubot does not have direct means to respond to your request because
it does not have write access to the databases where your data is
stored. Therefore, we will respond to your request by looping in
M-Lab and cooperating with them to implement such requests.
Specifically, if you request the erasure of restriction of your
personal data, Neubot will cooperate with M-Lab to anonymise your
data, thus taking it out of scope of the GDPR.

## Privacy by design and data minimisation when exercising your GDPR rights

Exercising the above-mentioned rights will require you to prove
your identity. Neubot is committed to data privacy and specifically
one of its core principles: data minimisation. As such, we will not
request a copy of your ID to prove your identity. Because we collect
only your IP address at the time the test is run, we have no means
of proving the test data belongs to you. Should you wish to exercise
the rights listed above, we require proof that this IP address
belonged to you at the point in time when you conducted a specific
test. Such proof could be provided in the form of a written
confirmation from the resource holder (‘owner’) of the IP block
your IP address is part of, for example your Internet Service
Provider. You can find the details of the resource holder by using
tools like [a WHOIS search](https://whois.net/).

## How to exercise your rights under the GDPR

Write to bassosimone@gmail.com formulating which right you wish to exercise
and in which capacity. If you wish to secure your message exchange via PGP
you may use the `738877AA6C829F26A431C5F480B691277733D95B` PGP key. (Because
Neubot does not have direct control over your data, it would be useful to
also loop into the discussion M-Lab since the beginnng.)

We encourage you to contact us at bassosimone@gmail.com if you
have a privacy related concern. However, you have the right to lodge
a complaint to a data protection authority of your choice if you
suspect us of improperly processing your personal data. You may
choose to do so with the data protection authority of the European
member state you live or work in but are free to turn to a data
protection authority of the member state of your choices. A complete
list of data protection authorities and their contact details in
the EU can be found on the website of the European data protection
board.

We regularly review this Privacy Policy and make sure that we process
the information we collect in compliance with it. Regardless of
where your information is processed, we apply the same practices
described in this Policy. When we receive formal written complaints,
we respond by contacting the person who made the complaint. We may
work with the appropriate regulatory authorities, including local
data protection authorities, to resolve any complaints regarding
the processing of data that we cannot resolve with you directly.
You can contact your local data protection authority if you have
concerns regarding your rights under local law.

## Changes to this Privacy Policy

We may modify this privacy policy at any time. The text is committed to
GitHub, so it will always be possble to inspect its history.
