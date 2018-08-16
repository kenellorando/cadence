#!/usr/bin/python

import socket
import select
import sys
import time
import calendar
import os
import hashlib
import base64
import math
import traceback
import pg8000
import string
import logging
import logging.handlers
from telnetlib import Telnet
from urllib import parse
from threading import Thread, current_thread
from configparser import ConfigParser

# Prep work
# Load in our configuration
# First, load in the defaults
defaultconfig = ConfigParser(interpolation=None)
defaultconfig.read('default-config.ini')

# Now use those defaults to load in the overrides
config = ConfigParser(defaults=defaultconfig['DEFAULT'], interpolation=None)
config.read('config.ini')
config = config['DEFAULT']

level = config['loglevel']
# Translate a log level (as configured) into a useful log level
leveldict = {
    "DEBUG"    : logging.DEBUG,
    "INFO"     : logging.INFO,
    "WARNING"  : logging.WARNING,
    "ERROR"    : logging.ERROR,
    "CRITICAL" : logging.CRITICAL
}

# Check if our log level is a valid name
if level in leveldict.keys():
    level = leveldict[level]
# If it isn't, try to convert it from an integer
else:
    try:
        level = int(level)
    except ValueError as ex:
        # We don't understand this level.
        # Raise a new exception with a somewhat more helpful error
        raise RuntimeError("Could not interpret \""+level+"\" as a valid logging level") from ex

# If logs directory does not exist, create it
logdir = os.path.join(os.path.dirname(os.path.abspath(__file__)), config['logdirectory'])
logdir = os.path.realpath(logdir)
if not os.path.exists(logdir):
    os.makedirs(logdir)

# Prepare the handlers for our logger (using the data as configured)
handlers = []
if config.getboolean('log_to_console'):
    # Find the correct stream and add it
    if config['logstream']=="stdout":
        handlers.append(logging.StreamHandler(sys.stdout))
    elif config['logstream']=="stderr":
        handlers.append(logging.StreamHandler(sys.stderr))
    else:
        # No found handler. Warn the user and use stderr
        from warnings import warn
        warn("Could not parse "+config['logstream']+" as a console stream. Using stderr.")
        handlers.append(logging.StreamHandler(sys.stderr))
if config.getboolean('log_to_disk'):
    # Assemble and add the timed rotating file handler
    handlers.append(logging.handlers.TimedRotatingFileHandler(os.path.join(logdir, config['logfile']), 'D', 1, 30))

# If the handlers list is empty, reconfigure so that logging uses a StreamHandler, but has an unreasonably high logging level.
# That way, logging will be effectively disabled, without adding any code anywhere else
if len(handlers)==0:
    handlers.append(logging.StreamHandler(sys.stdout))
    level=math.pow(2, 64)

# Log both to the console and to a daily rotating file, storing no more than 30 days of logs
logging.basicConfig(level=level,
                    format=config['logformat'],
                    handlers=handlers)
logger = logging.getLogger(config['logger'])
logger.setLevel(level)

port = int(sys.argv[1])
directory = os.path.realpath(sys.argv[2]).encode()

caching=0

# Check if we might have the -c flag
if len(sys.argv)>3 or config.getboolean('force_caching'):
    if config.getboolean('force_caching') or sys.argv[3].startswith("-c"):
        if len(sys.argv)>4:
            caching = int(sys.argv[4])
        else:
            caching = int(config['default_caching_duration'])
    else:
        logger.warning("Did not understand argument %s.", sys.argv[3])

sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

sock.bind(("", port))
sock.listen(int(config['backlog']))

# Set sock as nonblocking
sock.setblocking(False)

# Helper functions
def waitingRequest(s, blocksize=4096):
    "Returns a string containing one complete HTTP request from s, loaded in chunks of blocksize"

    out = s.recv(blocksize)
    # Hopefully, our headers will be in the first blocksize.
    # But first, we know that if output is smaller than blocksize, we have everything that's ready for us
    if len(out)<blocksize:
        logger.debug("Downloaded request of size %d.", len(out))
        return out

    # While true, try to parse a content size out of our received data, and if we can't, fetch a block.
    contentSize = 0
    while True:
        block = s.recv(blocksize)
        out += block
        for line in block.split("\r\n"):
            if line.startswith("Content-Length: "):
                contentSize=int(line.split(": ")[1])
                break # Only use the first content-length header
    # "Worst" case scenario is that Content-Length is the last header.
    # In that case, we'll have four more bytes (CRLFCRLF), then the content bytes
    contentSize += 4
    # Load the content into out
    while contentSize>blocksize:
        out += s.recv(blocksize)
        contentSize -= blocksize
    if contentSize>0:
        out += s.recv(contentSize)

    logger.debug("Downloaded request of size %d.", len(out))

    # out should now contain all of our request
    return out

def mimeTypeOf (filename):
    "Attempts to find the appropriate MIME type for this file by extension (MIME types taken from https://www.freeformatter.com/mime-types-list.html)"

    parts = filename.decode().split(".")
    if len(parts)<2:
        # The file has no extension.
        # Default to application/octet-stream
        logger.debug("Assumed file %s was type %s (no extension).", filename.decode(), "application/octet-stream")
        return "application/octet-stream"

    # The extension is whatever is after the last '.' in the filename
    # Switch to lowercase for comparison
    extension = parts[-1].lower()
    logger.debug("Extension %s", extension)

    # Giant dictionary of extensions -> MIME types
    # In order to keep this whole thing from being recreated every time a request for a file is made,
    #   only allocate it the first time the function is called and store it as a local for other requests
    if not hasattr(mimeTypeOf, "dictionary"):
        mimeTypeOf.dictionary = {
            "es": "application/ecmascript",
            "epub": "application/epub+zip",
            "jar": "application/java-archive",
            "class": "application/java-vm",
            "js": "application/javascript",
            "json": "application/json",
            "mathml": "application/mathml+xml",
            "mp4": "application/mp4",
            "doc": "application/msword",
            "bin": "application/octet-stream",
            "ogx": "application/ogg",
            "ogg": "application/ogg",
            "onetoc": "application/onenote",
            "pdf": "application/pdf",
            "ai": "application/postscript",
            "ps": "application/postscript",
            "rss": "application/rss+xml",
            "rtf": "application/rtf",
            "gram": "application/srgs",
            "sru": "application/sru+xml",
            "ssml": "application/ssml+xml",
            "tsd": "application/timestamped-data",
            "apk": "application/vnd.android.package-archive",
            "m3u8": "application/vnd.apple.mpegurl",
            "ppd": "application/vnd.cups-ppd",
            "gmx": "application/vnd.gmx",
            "xls": "application/vnd.ms.excel",
            "eot": "application/vnd.ms-fontobject",
            "chm": "application/vnd.ms-htmlhelp",
            "ppt": "application/vnd.ms-powerpoint",
            "mus": "application/vnd.musician",
            "odf": "application/vnd.oasis.opendocument.formula",
            "odg": "application/vnd.oasis.opendocument.graphics",
            "odi": "application/vnd.oasis.opendocument.image",
            "odp": "application/vnd.oasis.opendocument.presentation",
            "ods": "application/vnd.oasis.opendocument.spreadsheet",
            "odt": "application/vnd.oasis.opendocument.text",
            "pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
            "ppsx": "application/vnd.openxmlformats-officedocument.presentationml.slideshow",
            "xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
            "docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
            "rm": "application/vnd.rn-realmedia",
            "unityweb": "application/vnd.unity",
            "wpd": "application/vnd.wordperfect",
            "hlp": "application/winhlp",
            "7z": "application/x-7z-compressed",
            "dmg": "application/x-apple-diskimage",
            "bz": "application/x-bzip",
            "bz2": "application/x-bzip2",
            "vcd": "application/x-cdlink",
            "chat": "application/x-chat",
            "pgn": "application/x-chess-pgn",
            "csh": "application/x-csh",
            "deb": "application/x-debian-package",
            "wad": "application/x-doom",
            "dvi": "application/x-dvi",
            "otf": "application/x-font-otf",
            "pcf": "application/x-font-pcf",
            "ttf": "application/x-font-ttf",
            "pfa": "application/x-font-type1",
            "woff": "application/x-font-woff",
            "latex": "application/x-latex",
            "clp": "application/x-msclip",
            "exe": "application/x-msdownload",
            "pub": "application/x-mspublisher",
            "rar": "application/x-rar-compressed",
            "sh": "application/x-sh",
            "swf": "application/x-shockwave-flash",
            "xap": "application/x-silverlight-app",
            "tar": "application/x-tar",
            "tex": "application/x-tex",
            "texinfo": "application/x-texinfo",
            "xhtml": "application/xhtml+xml",
            "dtd": "application/xml+dtd",
            "zip": "application/zip",
            "mid": "audio/midi",
            "mp4a": "audio/mp4",
            "mpga": "audio/mpeg",
            "oga": "audio/ogg",
            "dts": "audio/vnd.dts",
            "dtshd": "audio/vnd.dts.hd",
            "weba": "audio/webm",
            "aac": "audio/x-aac",
            "m3u": "audio/x-mpegurl",
            "wma": "audio/x-ms-wma",
            "wav": "audio/x-wav",
            "bmp": "image/bmp",
            "gif": "image/gif",
            "jpg": "image/jpeg",
            "jpeg": "image/jpeg",
            "pjpeg": "image/pjpeg",
            "png": "image/png",
            "svg": "image/svg+xml",
            "tiff": "image/tiff",
            "psd": "image/vnd.adobe.photoshop",
            "sub": "image/vnd.dvb.subtitle",
            "webp": "image/webp",
            "ico": "image/x-icon",
            "pbm": "image/x-portable-bitmap",
            "eml": "message/rfc822",
            "ics": "text/calendar",
            "css": "text/css",
            "csv": "text/csv",
            "html": "text/html",
            "txt": "text/plain",
            "rtx": "text/richtext",
            "sgml": "text/sgml",
            "tsv": "text/tab-separated-values",
            "ttl": "text/turtle",
            "uri": "text/uri-list",
            "curl": "text/vnd.curl",
            "scurl": "text/vnd.curl.scurl",
            "s": "text/x-asm",
            "c": "text/x-c",
            "f": "text/x-fortran",
            "java": "text/x-java-source,java",
            "vcs": "text/x-vcalendar",
            "vcf": "text/x-vcard",
            "yaml": "text/yaml",
            "3gp": "video/3gpp",
            "3g2": "video/3gpp2",
            "h264": "video/h264",
            "jpgv": "video/jpeg",
            "mp4": "video/mp4",
            "mpeg": "video/mpeg",
            "ogv": "video/ogg",
            "qt": "video/quicktime",
            "mxu": "video/vnd.mpegurl",
            "webm": "video/webm",
            "f4v": "video/x-f4v",
            "flv": "video/x-flv",
            "m4v": "video/x-m4v",
            "wmv": "video/x-ms-wmv",
            "avi": "video/x-msvideo",
        }

    if not extension in mimeTypeOf.dictionary.keys():
        # We don't recognize this filetype
        # Default to application/octet-stream
        logger.debug("Assumed file %s was type %s (unknown extension).", filename.decode(), "application/octet-stream")
        return "application/octet-stream"

    # Recognized filetype. Return it.
    logger.debug("Guessed file %s was type %s.", filename.decode(), mimeTypeOf.dictionary[extension])
    return mimeTypeOf.dictionary[extension]

def requestBody(request):
    "Returns only the body of a given HTTP request"

    return request.partition(b"\r\n\r\n")[2].decode()

def ETag(content):
    "Returns a standard ETag of content (base64 encoded sha256) for use anywhere ETags are used in the server. ETags are returned in the same form content was passed (string or bytes)"

    if isinstance(content, str):
        return base64.urlsafe_b64encode(hashlib.sha256(content.encode()).digest()).decode()
    else:
        return base64.urlsafe_b64encode(hashlib.sha256(content).digest())

def HTTP_time(at=time.time()):
    "Returns a string formatted as an HTTP time, corresponding to the unix time specified by at (defaults to the present)"

    return time.strftime("%a, %d %b %Y %H:%M:%S GMT", time.gmtime(at))

def parse_HTTP_time(at):
    "Returns a Unix timestamp from an HTTP timestamp"

    if isinstance(at, bytes):
        at=at.decode()

    return calendar.timegm(time.strptime(at, "%a, %d %b %Y %H:%M:%S GMT"))

def basicHeaders(status, contentType):
    "Constructs and returns a basic set of headers for a response (Does not end the header block)"

    # For performance, pre-create a format string for basic headers (we use this function a lot)
    if not hasattr(basicHeaders, "format"):
        basicHeaders.format =  "HTTP/1.1 {0}\r\n"
        basicHeaders.format += "Date: {1}\r\n"
        basicHeaders.format += "Connection: close\r\n"

        # Advertise the configured state of our range request support
        if config.getboolean("enable_range_requests"):
            basicHeaders.format += "Accept-Ranges: bytes\r\n"
        else:
            basicHeaders.format += "Accept-Ranges: none\r\n"

        basicHeaders.format += "\r\n".join([s.strip() for s in config['additional_headers'].split(',')])+"\r\n"

        # Add cache-control header iff we have caching set
        if caching>0:
            basicHeaders.format += "Cache-Control: public, max-age="+str(caching)+"\r\n"

        basicHeaders.format += "Content-Type: {2}\r\n"

    # Format in our arguments and return
    return basicHeaders.format.format(status, HTTP_time(), contentType).encode()

def constructResponse(unendedHeaders, content, etag=None):
    "Attaches unendedHeaders and content into one HTTP response (adding content-length in the process), optionally overriding the etag"

    response =  unendedHeaders

    # Add ETag iff we have caching set
    if caching>0:
        # Either generate our own, or use the provided one
        if etag==None:
            response += b"ETag: \""+ETag(content)+b"\"\r\n"
        else:
            response += b"ETag: \""+etag+b"\"\r\n"

    response += b"Content-Length: "+str(len(content)).encode()+b"\r\n\r\n"
    if isinstance(content, str):
        response += content.encode()
    else:
        response += content
    return response

def queueResponse(sock, response):
    "Prepare the response to be sent on the socket sock. No work is done to response before send."

    openconn.append(Connection(sock, True, content=response))

def sendResponse(status, contentType, content, sock, headers=[], etag=None):
    "Constructs and sends a response with the first three parameters via sock, optionally with additional headers, and optionally overriding the ETag"

    # Attempt to handle unencoded content
    # This occasionally throws TypeErrors, for no reason I can tell.
    # It complains that 'str' object is not callable... But the only thing I'm calling is type?
    try:
        if isinstance(content, str):
            content=content.encode()
    except:
        logger.debug("Strange error while attempting str/bytes detection.", exc_info=True)

        # At least keep from crashing.
        content=bytes(content)

    # If additional headers are specified, format them for HTTP
    # Else, send as normal
    if len(headers)>0:
        queueResponse(sock, constructResponse(basicHeaders(status, contentType)+("\r\n".join(headers)+"\r\n").encode(), content, etag))
    else:
        queueResponse(sock, constructResponse(basicHeaders(status, contentType), content, etag))

    logger.info("Queued response for socket %d.", sock.fileno())
    logger.debug("Response had %d additional headers: \"%s\".", len(headers), ", ".join(headers))

# Probably won't see much use for this... But need it at least for 400 bad request
def generateErrorPage(title, description):
    "Returns the HTML for an error page with title and description"

    # For performance, construct this once the first time an error page is generated
    if not hasattr(generateErrorPage, "format"):
        generateErrorPage.format =  "<!DOCTYPE html>\n"
        generateErrorPage.format += "<html>\n"
        generateErrorPage.format += "  <head>\n"
        generateErrorPage.format += "    <title>{0}</title>\n"
        generateErrorPage.format += "  </head>\n"
        generateErrorPage.format += "  <body>\n"
        generateErrorPage.format += "    <h1 style='text-align: center; width:100%'>{0}</h1>\n"
        generateErrorPage.format += "    <p>{1}</p>\n"
        generateErrorPage.format += "  </body>\n"
        generateErrorPage.format += "</html>\n"

    # Use string formatting to insert the parameters into the page
    return generateErrorPage.format.format(title, description).encode()

def ariaSearch(requestBody, conn):
    "Performs the action of an ARIA search as specified in the body, sending results on the passed connection"

    if not hasattr(ariaSearch, "timeout"):
        # Pre-store certain values from configuration
        # Process all the timeout values we claimed to support
        ariaSearch.timeout=config['db_timeout']
        if ariaSearch.timeout=="None":
            ariaSearch.timeout=None
        else:
            ariaSearch.timeout=int(timeout)
            if ariaSearch.timeout<=0:
                ariaSearch.timeout=None

        # Incomplete database search query string
        ariaSearch.selectfrom="SELECT "+config['db_column_title']+", "+config['db_column_artist']+", "+config['db_column_path']+" FROM "+config['db_table']+" "

    # Accept either a socket or a Connection
    sock = conn
    # If conn is a Connection, use the socket from it
    # (we don't use any information from Connection)
    if type(conn) is Connection:
        sock = conn.conn

    # Log the search
    logger.info("Received a search request on socket %d.", sock.fileno())
    logger.debug("Search body was: %s.", requestBody)

    # Parse the query
    query = ""
    try:
        query = parse.parse_qs(requestBody)["search"][0]
    except KeyError:
        # Some wiseguy sent us a bad request. Tsk tsk.
        # Send an error message that the frontend will (should) ignore
        sendResponse("400 Bad Request", "application/json", "Invalid request - "+requestBody+" does not contain a search key.", sock, ["Warning: 199 Cadence \"Search request \'"+requestBody+"\' could not be parsed into a search term.\""])
        return

    # Attempt to connect to the database
    try:
        db = pg8000.connect(user=config['db_username'], host=config['db_host'],
                            port=int(config['db_port']), database=config['db_name'],
                            password=config['db_password'], ssl=config.getboolean('db_encrypt'),
                            timeout=ariaSearch.timeout)
        cursor = db.cursor()
    except:
        # Send the client an error message
        sendResponse("500 Internal Server Error", "application/json", "The server could not access the ARIA database.", sock)

        # Log the exception
        logger.exception("Could not connect to ARIA database.", exc_info=True)

    # Now, try to conduct the search using that connection
    try:
        results=[]
        q=query.lower()

        d = q.translate(str.maketrans(dict.fromkeys(string.punctuation)))

        # Check for our special query forms, and get results out of them
        if d.startswith("songs named "):
            cursor.execute(ariaSearch.selectfrom+"WHERE "+config['db_column_title']+" ILIKE %s", (q[12:],))
        elif d.startswith("songs by "):
            cursor.execute(ariaSearch.selectfrom+"WHERE "+config['db_column_artist']+" ILIKE %s", (q[9:],))
        elif d.endswith(" songs") and config['db_column_genre']!="None":
            cursor.execute(ariaSearch.selectfrom+"WHERE "+config['db_column_genre']+" ILIKE %s", (q[:-6],))
        else:
            # We don't have a special form.
            # For now, we haven't yet agreed on how the server should behave in this situation
            # But I'm sure it'll include results where the artist or title match the query.
            cursor.execute(ariaSearch.selectfrom+"WHERE "+config['db_column_artist']+" ILIKE %s OR "+config['db_column_title']+" ILIKE %s", (q, q))

        # Save our results
        results=cursor.fetchall()

        # Store the number of results for the log
        length=len(results)

        # Close the database connection and cursor
        db.close()
        cursor.close()

        # Now, we have a collection of results. We need to make it a JSON-parsable collection
        # In addition, we need to make sure it has the appropriate names for the ARIA frontend
        # First, let's do that second part, with a JSON-parsable formatted string
        # Making our lives more difficult is the fact that this data can technically contain quotes.
        # We need to escape those to not confuse the browser, with a simply disgusting replace call
        formatter="\"title\": \"{0}\", \"artist\": \"{1}\", \"path\": \"{2}\""
        results=[formatter.format(song[0].replace("\\", "\\\\").replace("\"", "\\\""),
                                  song[1].replace("\\", "\\\\").replace("\"", "\\\""),
                                  song[2].replace("\\", "\\\\").replace("\"", "\\\""))
                 for song in results]

        # If no results, just send an empty array
        if length==0:
            results="[]"
        # Otherwise perform normal JSON formatter
        else:
            # results is now a list of strings, each of which is an ARIA search result in JSON encoding
            # Now, join those strings together to make a single JSON string
            results="[{"+"},{".join(results)+"}]" # Disgusting, but surprisingly effective
            # Who needs JSON libraries anyway?

        # Send that result string to the user
        sendResponse("200 OK", "application/json", results, sock)

        # Log results
        logger.debug("Search for \"%s\" had %d results - %s.", query, length, results)
    except:
        # Well, we couldn't search. Tell the user and log the error

        # Send a response to the user
        sendResponse("500 Internal Server Error", "application/json", "The server could not query the database.", sock)

        # Log the error
        logger.exception("Could connect to database, but could not execute search.", exc_info=True)

def ariaRequest(requestBody, conn):
    "Performs the action of an ARIA search as specified in the body, sending results on the passed connection"

    # Setup for connection IP
    sock = conn
    # If connection, sock is the socket from the connection
    if type(conn) is Connection:
        sock = conn.conn
    # Otherwise, conn needs to be a socket.
    # That socket stays in sock
    # conn becomes a read Connection covering that socket with the IP set from the peername
    else:
        conn = Connection(sock, False)
        conn.IP = sock.getpeername()[0]

    # Log the request
    logger.info("Received a song request on socket %d.", sock.fileno())
    logger.debug("Request body was: %s.", requestBody)

    # We need a static variable to track last-request times per-user
    # Initialize it on first run to an empty array
    if not hasattr(ariaRequest, "timeouts"):
        ariaRequest.timeouts={}
        ariaRequest.timeoutSeconds=float(config['request_timeout'])

        # Data for special requests
        ariaRequest.specialEnabled=config.getboolean('special_request_timeouts')
        ariaRequest.specialForced=config.getboolean('special_request_force_enable')

        # If enabled, also add in our whitelist
        if (ariaRequest.specialEnabled):
            whitelist = config['special_request_whitelist']
            # If the whitelist is set to empty, disable special timeouts
            # (Unless they're force-enabled)
            if whitelist == "None" and not ariaRequest.specialForced:
                ariaRequest.specialEnabled = False
            # Else, parse out a list of addresses and save it
            else:
                whitelist = whitelist.split(',')
                ariaRequest.specialWhitelist = [addr.strip() for addr in whitelist]

        # Configure the blacklist
        if config['request_blacklist']=="None":
            ariaRequest.requestBlacklist=[]
        else:
            ariaRequest.requestBlacklist=[addr.strip() for addr in config['request_blacklist'].split(',')]

        # Configure the whitelist
        # Since the mechanism is intended to conditionally allow special requests, disable it if those are off.
        # Since the mechanism is only triggered when the blacklist is triggered, disable it if the blacklist is empty.
        if config['request_whitelist']=="None" or not ariaRequest.specialEnabled or ariaRequest.requestBlacklist==[]:
            ariaRequest.requestWhitelist=[]
        else:
            ariaRequest.requestWhitelist=[addr.strip() for addr in config['request_whitelist'].split(',')]

        # Configure the liquidsoap target
        ariaRequest.liquidsoapPort=int(config['liquidsoap_port'])

    request = parse.parse_qs(requestBody)

    # Either use client IP or provided tag for timeouts
    tag = conn.IP

    try:
        # If request timeout is zero, throw a KeyError to skip the timeout logic
        # This is a bit hacky, I know, but its easy, and I don't expect production instances to do this
        if ariaRequest.timeoutSeconds <= 0.0:
            raise KeyError("Fake error to skip timeout logic")

        # Check if config is set to let us try to use a tag
        if ariaRequest.specialEnabled:
            # Check if this client is on the whitelist allowed to use special request timeouts
            # Alternatively, check if we're allowing specials from all addresses
            if tag in ariaRequest.specialWhitelist or ariaRequest.specialForced:
                # We're in business.
                # Check if the request includes a tag
                if "tag" in request.keys():
                    # Use that tag for timeouts (joining any possible additional values with ampersands)
                    tag += '/' + '&'.join(request["tag"])

        # Check if the user is blacklisted
        if tag not in ariaRequest.requestWhitelist and (tag in ariaRequest.requestBlacklist or conn.IP in ariaRequest.requestBlacklist):
            # User on the blacklist. Issue a Forbidden response.
            sendResponse("403 Forbidden",
                         "text/plain",
                         "ARIA: The server administrator has forbidden you from submitting requests.",
                         sock,
                         ["Warning: 299 Cadence The server has been configured to block this user from requesting songs."])

            # Log the blacklist error.
            logger.warning("User with tag %s at address %s is on the request blacklist, and therefore was bocked from making a request.", tag, conn.IP)
            return

        timeout=ariaRequest.timeouts[tag]
        logger.debug("Request timeout for %s at second %f. Current time %f.", tag, timeout+ariaRequest.timeoutSeconds, time.monotonic())
        if timeout+ariaRequest.timeoutSeconds>time.monotonic():
            # Timeout period hasn't passed yet. Return an error message (actually, the same message the Node.js server used)
            # Since we're so nice, we'll even send a header telling the client how long is left on the timeout. Most clients won't even look for it, but we do provide the information.
            sendResponse("429 Too Many Requests",
                         "text/plain",
                         "ARIA: Request rejected, you must wait five minutes between requests.",
                         sock,
                         ["Retry-After: "+str(math.ceil((timeout+ariaRequest.timeoutSeconds)-time.monotonic()))])

            # Log timeout
            logger.info("Request too close to previous request from user %s.", tag)
            return
    except KeyError:
        pass

    # If we get here, the timeout mechanism is allowing the request.
    # First, isolate the path of our request
    path = request["path"][0]
    logger.info("Path: %s", path)

    # Use telnet to connect to the stream client and transmit the request
    connection = Telnet(config['liquidsoap_host'], ariaRequest.liquidsoapPort)
    try:
        connection.write(("request.push "+path).encode())
        response=connection.read_until(b'END', 2).decode()

        logger.info("Pushed request. Source client response: %s", response)

        # And now update the timeout for this user if the timeout is positive
        if ariaRequest.timeoutSeconds>0:
            ariaRequest.timeouts[tag]=time.monotonic()
            logger.debug("Updated timeout: User at %s may request again at %f.", tag, time.monotonic()+ariaRequest.timeoutSeconds)

        # Inform the user that their request has been received.
        # Include a custom header with the queue position.
        # Provide the same information in a comment in the ariaSays element.
        pos = ""
        try:
            pos = str(int(response[:-3])) # read_until includes the "END", so we have to strip it out.
        except:
            pos = "Unknown"

        sendResponse("200 OK",
                     "text/html",
                     "ARIA: Request received!\n"+
                     "<!-- Position in queue: "+pos+" -->",
                     sock,
                     ["X-Queue-Position: "+pos])
    except:
        logger.exception("Exception while requesting song %s.", path, exc_info=True)

        # Something bad happened while contacting the stream client
        # We'll tell the client that the request service is unavailable, until September 1 2018
        sendResponse("503 Service Unavailable",
                     "text/html",
                     "ARIA: Something went wrong while processing your request.",
                     sock,
                     ["Retry-After: Sat, 01 Sep 2018 00:00:00 GMT"])
    finally:
        connection.close()

# Class to store an open connection
class Connection:
    def __init__(self, conn, isWrite, isAccept=False, content=None, IP=None):
        self.conn = conn
        self.isWrite=isWrite
        self.isAccept = isAccept
        self.content = content
        self.IP=IP

    # Follows configured behavior to attempt to get an IP out of request headers
    def setIPFrom(self, requestHeaders):
        try:
            header = config['client_identification_header']
            if header == "None":
                # Use the socket connection address and return
                self.IP=self.conn.getpeername()[0]
                return

            # Attempt to find that header in the request headers
            # requestHeaders can either be a list of headers, bytes, or a string representing the headers
            # Either way, make sure it ends up as a list of strings
            lines = []
            if type(requestHeaders) is list:
                if type(requestHeaders[0]) is bytes:
                    lines = [line.decode() for line in requestHeaders]
                else:
                    # Assume strings
                    lines = requestHeaders
            elif type(requestHeaders) is str:
                lines = requestHeaders.split("\r\n")
            elif type(requestHeaders) is bytes:
                lines = requestHeaders.decode().split("\r\n")
            else:
                # Assume it's some sort of collection
                lines = [line for line in requestHeaders]

            # Lines is now a list of strings, where each string is an HTTP header.
            for line in lines:
                if line.startswith(header):
                    # We've found our header.
                    vals = line.partition(": ")
                    value = vals[2]

                    # Handle standard headers which require more processing
                    if vals[0]=="X-Forwarded-For":
                        # X-Forwarded-For includes a list of forwarding proxies we don't care about.
                        value=value.partition(",")[0]
                    elif vals[0]=="Forwarded":
                        # Forwarded includes more data than just identifier
                        parts=[part.strip() for part in value.split(';')]

                        # Get the field that contains source data
                        for part in parts:
                            if part[:4].lower() == "for=":
                                # Part is our part.
                                value=part[4:]
                                break

                        # We're probably done, unless the address is IPv6.
                        # IPv6 records, for no apparent reason, must be in quotes and brackets. Strip both just in case
                        if value.startswith('\"'):
                            value=value.strip('[]\"')

                    # Set our IP to be that value
                    self.IP=value
                    return

            # We didn't find the header. Fall back to the socket connection address.
            self.IP=self.conn.getpeername()[0]
        except OSError:
            logger.exception("Exception while attempting to read client IP from socket %d.", self.conn.fileno(), exc_info=True)
            self.IP="Unknown (exception while processing IP address; See log for socket "+str(self.conn.fileno())+")."
            # Note, happily, that including the socket number in there at least provides some separation between connections.

    # For compatibility with select
    def fileno(self):
        return self.conn.fileno()

# List of open connections
openconn = []

# Pre-create mimeTypeOf dictionary, basic headers, and error page data
mimeTypeOf(b"MimeType.precreate.file")
generateErrorPage("PRECREATION", "YOU SHOULD NEVER SEE THIS")
basicHeaders("599 Server Pre-create", "MimeType/precreate.file")

# Network operation helper functions
def readFrom(read, log=True):
    "Performs the operation of reading from the given Connection or set of Connections"

    # Log which thread we're on
    if log:
        logger.debug("Beginning read(s) on thread %s.", current_thread().name)

    # If this isn't a Connection, assume it's a collection of Connections and recurse
    if type(read) is not Connection:
        for r in read:
            readFrom(r, False)
        return

    # Ignore erroneous sockets (those with negative file descriptors)
    if read.fileno() < 0:
        # Drop the connection from openconn, close the error, and continue on our way
        # Ignore errors: What matters is that we don't do anything with the sockets
        try:
            openconn.remove(read)
            read.conn.close()
        except:
            pass
        return

    # For the accept socket, accept the connection and add it to the list
    if read.isAccept:
        # Accept as many connections as we can until none are immediately ready for accept
        try:
            while True:
                conn = read.conn.accept()[0]
                openconn.append(Connection(conn, False))
                logger.info("Accepting a new connection, attached socket %d.", conn.fileno())
        except socket.timeout:
            pass
        except BlockingIOError:
            pass
    else:
        logger.info("Processing request from socket %d.", read.fileno())
        # Fetch the HTTP request waiting on read
        request = waitingRequest(read.conn, int(config['HTTP_blocksize']))

        # Set the IP on the connection
        read.setIPFrom(request.partition(b"\r\n\r\n")[0])

        # If the request is zero-length, the client disconnected. Skip the work of figuring that out the hard way, and the unhelpful log message.
        # Log a better message, remove the connection from the list, and close the socket (skipping the rest of the loop)
        if len(request) == 0:
            logger.info("Empty request on socket %d.", read.fileno())
            openconn.remove(read)
            sendResponse("400 Bad Request",
                         "text/html",
                         generateErrorPage("400 Bad Request",
                                           "Your browser send an empty request."),
                         read.conn)
            return

        # Lines of the HTTP request (needed to read the header)
        lines = request.partition(b"\r\n\r\n")[0].split(b"\r\n")

        # The first line tells us what we're doing
        # If it's GET, we return the file specified via commandline
        # If it's HEAD, we return the headers we'd return for that file
        # If it's something else, return 405 Method Not Allowed
        method = lines[0]
        logger.debug("Method line %s", method.decode())
        if method.startswith(b"POST") and config.getboolean('enable_aria'):
            logger.info("Received POST request to %s.", method.split(b' ')[1].decode())
            if method.split(b' ')[1]==b"/search":
                Thread(target=ariaSearch, args=(requestBody(request), read)).start()
            elif method.split(b' ')[1]==b"/request":
                Thread(target=ariaRequest, args=(requestBody(request), read)).start()
            else:
                # No other paths can receive a POST.
                # Tell the browser it can't do that, and inform it that it may only use GET or HEAD here.
                sendResponse("405 Method Not Allowed",
                             "text/html",
                             generateErrorPage("405 Method Not Allowed",
                                               "Your browser attempted to perform an action the server doesn't support at this location."),
                             read.conn,
                             ["Allow: GET, HEAD"])

                # Log method not allowed
                logger.info("Issued method not allowed.")

            # No matter what, we've handled the request however we chose to.
            # Remove it from openconn
            openconn.remove(read)
            return
        elif not (method.startswith(b"GET") or method.startswith(b"HEAD")):
            # This server can't do anything with these methods.
            # So just tell the browser it's an invalid request
            sendResponse("501 Not Implemented",
                         "text/html",
                         generateErrorPage("501 Not Implemented",
                                           "Your browser sent a request to perform an action the server doesn't support."),
                         read.conn,
                         ["Allow: GET, HEAD"])
            openconn.remove(read)

            # Print note on error
            logger.info("Could not execute method %s.", method.decode())
            return

        # Parse the filename out of the request
        # Trim leading slashes to keep Python from thinking that the method refers to the root directory.
        filename = os.path.join(directory, method.split(b' ')[1].lstrip(b'/'))
        dir = False
        # If the filename is a directory, join it to "index.html"
        if os.path.isdir(filename):
            dir = True
            filename = os.path.join(filename, b"index.html")

        # Normalize the file path
        filename = os.path.realpath(filename)

        # Check if the relative path between the file and the service directory includes '..'
        # In other words, if one has to go 'up' in the directory structure to get to the target
        # If this is the case, return an error forbidding access to that file
        if b".." in os.path.relpath(filename, directory):
            # Detected attempt to access file outside allowed directory.
            # ACCESS DENIED
            sendResponse("403 Forbidden",
                         "text/html",
                         generateErrorPage("403 Forbidden",
                                           "You are not permitted to access \""+method.decode().split(' ')[1]+"\" on this server."),
                         read.conn,
                         ["Warning: 299 Cadence Access to files above the root directory of the served path is forbidden. This incident has been logged."])

            # Log an error, pertaining to the fact that an attempt to access forbidden data has been thwarted.
            logger.error("Client at %s attempted to access forbidden file %s, but was denied access.", read.IP, filename.decode())

            # Remove the read connection and continue
            openconn.remove(read)
            return

        # Perform redirect of directories that don't end in a separator or slash
        targ = method.split(b' ')[1].decode()
        if dir and not (targ.endswith(os.path.sep) or targ.endswith('/')):
            sendResponse("301 Moved Permanently",
                         "text/html",
                         b"",
                         read.conn,
                         ["Location: "+targ+'/'])

            # Log redirect
            logger.info("Issued redirect from %s to %s/.", targ, targ)

            # Remove read connection and continue
            openconn.remove(read)
            return

        # Guess the MIME type of the file.
        mimetype = mimeTypeOf(filename)

        # Read the file into memory
        logger.info("Attempting file read on file %s.", filename.decode())
        file = ""
        try:
            with open(filename, 'rb', 0) as f:
                file = f.read()
        except FileNotFoundError:
            # The file wasn't found.
            # Check for the 418 easter egg
            if method.split(b' ')[1].endswith(b"coffee") and config.getboolean('enable_418'):
                # Someone must be trying to get some coffee!
                # Too bad for them.
                # Image is, unsurprisingly, a teapot I rendered
                image = ""
                try:
                    with open(os.path.join(os.path.dirname(os.path.abspath(__file__)), "teapot.png"), 'rb', 0) as f:
                        image = base64.b64encode(f.read()).decode()
                except:
                    pass

                # If file load failed, just skip the image
                if len(image)==0:
                    sendResponse("418 I'm a teapot",
                                 "text/html",
                                 generateErrorPage("418 I'm a teapot",
                                                   "I'm sorry - I can't make coffee for you.<br>I'm a teapot."),
                                 read.conn)
                else:
                    sendResponse("418 I'm a teapot",
                                 "text/html",
                                 generateErrorPage("418 I'm a teapot",
                                                   "I'm sorry - I can't make coffee for you.</p>"+
                                                   "<img src=\"data:image/png;base64,"+image+"\" width=256 height=256><p>I'm a teapot."),
                                 read.conn)

                # Log the teapot
                logger.warning("Became a teapot in response to request for unfound file %s.", filename.decode())

                # Remove read connection and continue
                openconn.remove(read)
                file = ""

            # Not a teapot
            else:
                # Return 404.
                sendResponse("404 Not Found",
                             "text/html",
                             generateErrorPage("404 Not Found",
                                               "The requested file \""+method.decode().split(' ')[1]+
                                               "\" was not found on this server."),
                             read.conn)
                # Remove read connection and continue
                openconn.remove(read)

                # Print note on error
                logger.warning("Could not find file %s.", filename.decode())
                file = ""
        except:
            # Some unknown error occurred. Return 500.
            # First, generate our error message
            exc_type, exc_value, exc_traceback = sys.exc_info()
            message = ''.join(traceback.format_exception(exc_type, exc_value, exc_traceback))

            # Now send the error message.
            sendResponse("500 Internal Server Error",
                         "text/html",
                         generateErrorPage("500 Internal Server Error",
                                           "The server encountered an error while attempting to process your request.\n"+
                                           "<!-- Ok, since you know what you're doing, I'll confess.\n"+
                                           "I know what the error is. Python says:\n"+
                                           message+'\n'+" -->"),
                         read.conn)
            # Remove read connection and continue
            openconn.remove(read)

            # Print note on error
            logger.exception("Could not open file %s.", filename.decode(), exc_info=True)
            file = ""

        if file=="":
            logger.debug("Breaking off connection attempt due to file open issue.")
            return

        # Serve the file back to the client.
        # First, handle caching
        if caching>0:
            logger.debug("Caching is enabled, checking for If-None-Match")
            Etag = ""
            for line in lines:
                if line.startswith(b"If-None-Match: "):
                    Etag = line.split(b"\"")[1]
                    logger.debug("Found header - ETag %s.", Etag.decode())

            # If there was no If-None-Match, check for a provided If-Modified-Since
            if Etag == "":
                logger.debug("Found no ETag, searching for last modified time.")
                mtime = float("nan")
                for line in lines:
                    if line.startswith(b"If-Modified-Since: "):
                        mt = line.partition(b": ")[2]
                        mtime = parse_HTTP_time(mt)
                        logger.debug("Found header - mtime %f, from timestamp %s.", mtime, mt.decode())

                if mtime>=math.floor(os.path.getmtime(filename)):
                    # Last modified time was given (all NaN comparisons return false), and the file has not since been modified.
                    # Return basic headers, plus ETag and mtime
                    queueResponse(read.conn, basicHeaders("304 Not Modified", mimetype)+b"ETag: \""+ETag(file)+b"\"\r\nLast-Modified: "+HTTP_time(os.path.getmtime(filename)).encode()+b"\r\n\r\n")
                    logger.info("Client already has this file (not modified since %f [which is %s]).", mtime, HTTP_time(mtime))

                    # Remove read connection and move on.
                    openconn.remove(read)
                    return

            # If we have an ETag and it matches our file, return 304 Not Modified
            elif Etag == ETag(file):
                # ETag matches. Return our basic headers, plus the ETag and mtime
                queueResponse(read.conn, basicHeaders("304 Not Modified", mimetype)+b"ETag: \""+Etag+b"\"\r\nLast-Modified: "+HTTP_time(os.path.getmtime(filename)).encode()+b"\r\n\r\n")
                logger.info("Client already has this file (matching hash %s) - Issued 304.", Etag.decode())

                # Remove read connection and move on.
                openconn.remove(read)
                return

        # Check if we're doing a byte reply
        done=False
        for line in lines:
            # If we're not processing byte replies, break out of the loop
            # (this is here to reduce indentation on this really big loop)
            if not config.getboolean("enable_range_requests"):
                break

            if line.startswith(b"Range: "):
                # We have a byte-range-request
                logger.debug("Request on socket %d is a range request.", read.fileno())

                # Check for If-Range
                exit=False
                for l in lines:
                    if l.startswith(b"If-Range: "):
                        # We found an If-Range
                        value=l.partition(b": ")[2]

                        # Check if the If-Range is a last-modified or an ETag
                        # Because our ETags are base64 encoded, we can check for the presence of a space to do this
                        if b' ' in value:
                            # Value is a last-modified
                            mtime=parse_HTTP_time(value)
                            logger.debug("Request is using mtime for If-Range.")

                            # Compare mtimes
                            if mtime<math.floor(os.path.getmtime(filename)):
                                # The file has been modified. We have to do a full-file.
                                exit=True
                                logger.debug("File modified since %d (mtime %d).", mtime, math.floor(os.path.getmtime(filename)))

                            # Either way, we're done. Break out.
                            break

                        else:
                            # Value is an ETag
                            value=value.strip(b"\"")
                            logger.debug("Request is using ETag for If-Range.")

                            # Compare ETags
                            etag=ETag(file)
                            if value!=etag:
                                # The file has been modified. We have to do a full-file.
                                exit=True
                                logger.debug("File modified. Client ETag \"%s\", server ETag \"%s\".", value, etag)

                            # Either way, we're done. Break out.
                            break

                # If the If-Range says not to perform a byte-range reply, break out of the loop early
                if exit:
                    break

                # Perform a byte-range reply
                range=line.partition(b": ")[2]
                # 'Range' should look like "bytes=x-y"
                # Clip out those first six characters
                range=range[6:]
                # Now, trim our file to match that range, saving the original length and ETag
                # Catch errors in the process and treat them as being ill-formed
                # This includes multipart requests, which are currently considered more trouble than they're worth.
                try:
                    points=[int(point) if len(point)!=0 else None for point in range.partition(b"-")[::2]]
                except:
                    points=[-1, -1]

                    # Log that there was an exception
                    logger.exception("Exception while processing range request for %s. If this is a multipart request, consider submitting an issue on github to add support for your use-case.", line.partition(b": ")[2].decode(), exc_info=True)

                length=len(file)
                # Handle empty points
                if points[0]==None:
                    points[0]=0
                if points[1]==None:
                    points[1]=length-1

                if points[0]<0 or points[1]>length or points[0]>points[1]:
                    # The request cannot be satisfied
                    # (The request doesn't ask for a valid part of the file)
                    # Issue a 416
                    sendResponse("416 Range Not Satisfiable",
                                 "text/html",
                                 generateErrorPage("416 Range Not Satisfiable",
                                                   "The server was unable to satisfy your request for bytes {0} to {1} of a {2} byte file.".format(points[0], points[1], length)),
                                 read.conn,
                                 ["Content-Range: */"+str(length)])

                    # Log the problem
                    logger.warning("Could not satisfy request from socket %d for bytes %d to %d of %d byte file %s.", read.fileno(), points[0], points[1], length, filename)

                    # Remove read connection and continue
                    openconn.remove(read)
                    done=True
                    break

                etag=ETag(file)
                file=file[points[0]:points[1]+1]
                # File now only contain the range that was requested.
                # Send it off, with a Content-Range header explaining how much we sent.
                # Respect both GET and HEAD
                # Pass the ETag we calculated
                if method.startswith(b"GET"):
                    sendResponse("206 Partial Content",
                                 mimetype,
                                 file,
                                 read.conn,
                                 ["Content-Range: bytes {0}-{1}/{2}".format(points[0], points[1], length),
                                  "Last-Modified: "+HTTP_time(os.path.getmtime(filename))],
                                 etag)
                else:
                    queueResponse(read.conn, constructResponse(basicHeaders("206 Partial Content",
                                                                            mimetype)+
                                                               b"Last-Modified: "+HTTP_time(os.path.getmtime(filename)).encode()+b"\r\n"+
                                                               "Content-Range: bytes {0}-{1}/{2}\r\n".format(points[0], points[1], length).encode(),
                                                               etag).partition(b"\r\n\r\n")[0]+b"\r\n\r\n")
                    logger.info("Sent headers for partial request to socket %d.", read.fileno())

                # Now, remove read connection and move on
                openconn.remove(read)
                done=True
                break

        # Skip the normal full-file processing if we already sent a message
        if done:
            return

        # If we're here, we're not doing a byte range reply
        # If the method is GET, use sendResponse to send the file contents.
        if method.startswith(b"GET"):
            sendResponse("200 OK", mimetype, file, read.conn, ["Last-Modified: "+HTTP_time(os.path.getmtime(filename))])
        # If the method is HEAD, generate the same response, but strip the body
        else:
            queueResponse(read.conn, constructResponse(basicHeaders("200 OK", mimetype)+b"Last-Modified: "+HTTP_time(os.path.getmtime(filename)).encode()+b"\r\n", file).partition(b"\r\n\r\n")[0]+b"\r\n\r\n")
            logger.info("Sent headers to socket %d.", read.fileno())

        # Now that we're done, remove read connection and move on.
        openconn.remove(read)

def writeTo(write, log=True):
    "Performs the operation of writing to the given Connection or set of Connections"

    # Log which thread we're on
    if log:
        logger.debug("Beginning write(s) on thread %s.", current_thread().name)

    # If write isn't a Connection, assume it's a collection of Connections
    if type(write) is not Connection:
        for w in write:
            writeTo(w)
        return

    # Handling writes is a lot easier than reads, because the read logic has made all the decisions.
    write.conn.sendall(write.content)
    logger.info("Sent response to socket %d.", write.fileno())

    # Close the connection and remove it from the waiting list
    write.conn.close()
    openconn.remove(write)

def createThread(target, name, args):
    "Wrapper for the Thread constructor which takes positional arguments"

    return Thread(target=target, name=name, args=args)

def constantIterable(const):
    "A generator which always returns const"

    while True:
        yield const

def nameIterable(prefix):
    "A generator which generates an infinite sequence of strings, as prefix+id, for id in {0...infinity}"

    prefix=str(prefix)
    ID=0
    while True:
        yield prefix+str(ID)
        ID+=1

def splitInto(arr, n):
    "Splits arr into n roughly equally sized pieces."

    # See how we can divide the length of the array into n pieces
    quotient, remainder=divmod(len(arr), n)
    # Use some neat math and our divisions to split the array in a generator statement
    return (arr[i*quotient+min(i, remainder) : (i+1)*quotient+min(i+1, remainder)] for i in range(n))

maxThreads=int(config['max_threads'])
# Generators for thread creation maps
reader = constantIterable(readFrom)
writer = constantIterable(writeTo)
readname = nameIterable("reader")
writename = nameIterable("writer")

# Infinite loop for connection service
while True:
    # List of sockets we're waiting to read from or write to
    logger.debug("Assembling socket list")
    r = []
    w = []
    # Add all waiting connections
    for conn in openconn:
        # Either append to w or r depending on whether the socket is waiting for write or for read
        if conn.isWrite:
            w.append(conn)
        else:
            r.append(conn)
    # And also add the incoming connection accept socket
    r.append(Connection(sock, False, True))
    # Now, select sockets to process

    logger.debug("Selection...")
    readable, writeable, u2 = select.select(r, w, [])

    # If we're in single-thread mode
    if maxThreads==0:
        # Read from the readable sockets
        logger.debug("Selected %d readable sockets.", len(readable))
        for read in readable:
            readFrom(read)

        # Now, handle the writeable sockets
        logger.debug("Selected %d writeable sockets.", len(writeable))
        for write in writeable:
            writeTo(write)

    # We're performing operations on multiple threads
    # If the maximum number of threads is one, skip over the logic for splitting up the socket arrays
    elif maxThreads==1:
        # Read from the readable sockets in a read thread
        logger.debug("Selected %d readable sockets.", len(readable))
        reader = Thread(target=readFrom, name=next(readname), args=(readable,))
        reader.start()

        # Now, handle the writeable sockets in a write thread
        logger.debug("Selected %d writeable sockets.", len(writeable))
        writer = Thread(target=writeTo, name=next(writename), args=(writeable,))
        writer.start()

        # Wait for both threads to finish
        reader.join()
        writer.join()

    # We have to use multiple threads per operation
    else:
        logger.debug("Selected %d readable sockets.", len(readable))
        # Split up the readable sockets and read from them
        readers=[]
        # Our work pools start as one socket to one thread
        rpools=readable

        # If we don't have enough threads for that, split the work up into maxThreads pools
        if maxThreads>len(readable):
            rpools=splitInto(readable, maxThreads)

        # Create a list of threads to run reads on
        readers=list(map(createThread, reader, readname, ((read,) for read in rpools)))
        # ...and start all of those threads
        for thread in readers:
            thread.start()

        logger.debug("Selected %d writeable sockets.", len(writeable))
        # Split up the writeable sockets and write to them
        writers=[]
        # Our work pools start as one socket to one thread
        wpools=writeable

        # If we don't have enough threads for that, split the work up into maxThreads pools
        if maxThreads>len(writeable):
            wpools=splitInto(writeable, maxThreads)

        # Create a list of threads to run writes on
        writers=list(map(createThread, writer, writename, ((write,) for write in wpools)))
        # ...and start all of those threads
        for thread in writers:
            thread.start()

        # By here, all of our readers and writers are running.
        # Wait for all of them to end before returning to selection
        for r in readers:
            r.join()
        for w in writers:
            w.join()
