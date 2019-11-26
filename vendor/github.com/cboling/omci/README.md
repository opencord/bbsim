# OMCI

OMCI gopacket library supports the encoding and decoding of ITU G.988 OMCI
messages. It is currently a work in progress.

## Message Types supported and under unit test
The following OMCI message types currently have been coded and are covered
satisfactory by unit tests.

 - CreateRequest
 - CreateResponse
 - DeleteRequest
 - DeleteResponse
 - SetRequest
 - GetRequest
 - GetAllAlarmsRequest
 - GetAllAlarmsResponse
 - GetAllAlarmsNextRequest
 - MibUploadRequest
 - MibUploadResponse
 - MibUploadNextRequest
 - MibResetRequest
 - MibResetResponse
 - SynchronizeTimeRequest

## Message Types supported but lacking full unit test
The following OMCI message types currently have been coded and are partially covered
by unit tests, but work still remains for sufficient/better unit test coverage.

 - SetResponse
 - GetResponse
 - GetAllAlarmsNextResponse
 - MibUploadNextResponse
 - SynchronizeTimeResponse
 - AttributeValueChange
 - RebootRequest
 - RebootResponse
 - StartSoftwareDownloadRequest
 - GetNextRequest
 - GetNextResponse

## Message Types supported but lacking any unit test
The following OMCI message types currently have been coded but do not
have any unit test coverage.

 - StartSoftwareDownloadResponse
 - DownloadSectionRequest
 - DownloadSectionResponse
 - EndSoftwareDownloadRequest
 - EndSoftwareDownloadResponse
 - ActivateSoftwareRequest
 - ActivateSoftwareResponse
 - CommitSoftwareRequest
 - CommitSoftwareResponse
 - GetCurrentDataRequest
 - GetCurrentDataResponse
 - AlarmNotification
 
## Message Types not yet supported

The following OMCI message types currently have not been coded.

 - TestResult
 - TestRequest
 - TestResponse
 - SetTableRequest
 - SetTableResponse

## Current user-test coverage


## Other outstanding items

Besides OMCI Message decode/serialization, and associated unit tests, the following items
would be needed or useful in a first official release of this library. Some changes are
to be done in the generated OMCI ME code as well.

 - Specific examples of how to use this library (expand upon DecodeEncode.go examples)
   Include unknown ME examples and how to catch various common or expected errors
 - Add Alarm Table Support (generated MEs also)
 - Add AVC flag for appropriate attributes
 - For serialization, check early for message size exceeded
 - Check proper gopacket use of Payload/Contents properties and make sure we
   follow guidelines (if there are any)
 - For 'mebase.go' string output, look up ME name and output as needed
 - Look through 'error' messages and see if there are a few very common ones that
   could be moved to a custom class to allow for better user interception/decode of
   these errors.
 
The following would be 'nice' to have but are not necessary for initial code release
 - Extended message support
 - MIC Encode/Decode support

## Create requests
Currently the OMCI parser does not decode the default Set-By-Create attribute values
from the ITU document. So for attributes that do not have a default of zero, you must
specify the defaults if you use the 'meframe.go' routine to create a CreateRequest
for a specific Managed Entity Instance.
