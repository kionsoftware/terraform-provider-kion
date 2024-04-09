from json import JSONDecodeError
import requests
from constants import HEADERS
from constants import MAX_UNAUTH_RETRIES
from constants import UNAUTH_RETRY_COUNTER
from config import ARGS


def api_call(url, method='get', payload=None, headers=None, timeout=30, test=False):
    """
    API Call

    Common helper function for making the API calls needed for this script.

    Params:
        url         (str)   - full URL to call
        method      (str)   - API method - GET or POST
        payload     (dict)  - payload for POST requests
        headers     (dict)  - different headers to use
        timeout     (int)   - timeout for the call, defaults to 10
        test        (bool)  - if true, just test success of response and return
                              True / False accordingly, rather than returning the response data

    Return:
        success - response['data']
        failure - False
    """
    # check for the skip_ssl_verify flag
    if ARGS.skip_ssl_verify:
        verify = False
    else:
        verify = True

    # override headers if set
    if headers:
        _headers = headers
    else:
        _headers = HEADERS

    # make the API call without JSON decoding
    try:
        if method.lower() == 'get':
            response = requests.get(
                url, headers=_headers, timeout=timeout, verify=verify)
        elif method.lower() == 'post':
            if payload:
                response = requests.post(
                    url, headers=_headers, json=payload, timeout=timeout, verify=verify)
            else:
                response = requests.post(
                    url, headers=_headers, timeout=timeout, verify=verify)
        else:
            print("Unhandled method supplied to api_call function: %s" %
                  method.lower())
            return False
    except (requests.ConnectionError, requests.exceptions.ReadTimeout, requests.exceptions.Timeout) as e:
        print("Request to %s timed out. Error: %s" % (url, e))
        return False
    except requests.exceptions.TooManyRedirects as e:
        print("Connection to %s returned Too Many Redirects error: %s" % (url, e))
        return False
    except requests.exceptions.RequestException as e:
        print("Connection to %s resulted in error: %s" % (url, e))
        return False
    except Exception as e:
        print("Exception occurred during connection to %s: %s" % (url, e))
        return False
    else:

        # at this point, no exceptions were thrown so the
        # the request succeeded

        # check if test is True, if so return True
        if test:
            return True

        # test for valid json response
        try:
            response.json()
        except JSONDecodeError as e:
            print("JSON decode error on response: %s, %s" % (response, e))
            return False
        else:
            response = response.json()

        if response['status'] == 200:
            # reset the unauth retry counter
            global UNAUTH_RETRY_COUNTER
            UNAUTH_RETRY_COUNTER = 0
            return response['data']
        elif response['status'] == 201:
            # 201's are the return code for resource creations
            # and the response object can vary, so just return the whole thing
            # and make the calling function deal with it
            return response
        elif response['status'] == 401:
            # retry up to MAX_UNAUTH_RETRIES
            if UNAUTH_RETRY_COUNTER < MAX_UNAUTH_RETRIES:
                retries = MAX_UNAUTH_RETRIES - UNAUTH_RETRY_COUNTER
                print(
                    "Received unauthorized response. Will retry %s more times." % retries)
                UNAUTH_RETRY_COUNTER += 1
                api_call(url)
            else:
                print("Hit max unauth retries.")
                return False
        else:
            print(response['status'])
            print("Error calling API: %s\n%s" % (url, response))
        return False
