#!/usr/bin/python

import socket
import select
import sys
import time
import os
import hashlib

# Prep work
port = int(sys.argv[1])
directory = sys.argv[2]

caching=0

# Check if we might have the -c flag
if len(sys.argv)>3:
    if sys.argv[3].startswith("-c"):
        if len(sys.argv)>4:
            caching = int(sys.argv[4])
        else:
            caching = 3600 # One hour caching by default
    else:
        print "Warning: Did not understand argument "+sys.argv[3]

sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

sock.bind(("", port))
sock.listen(5))

# Helper functions
def waitingRequest(s, blocksize=4096):
    "Returns a string containing one complete HTTP request from s, loaded in chunks of blocksize"

    out = s.recv(blocksize)
    # Hopefully, our headers will be in the first blocksize.
    # But first, we know that if output is smaller than blocksize, we have everything that's ready for us
    if len(out)<blocksize:
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

    # out should now contain all of our request
    return out

def mimeTypeOf (filename):
    "Attempts to find the appropriate MIME type for this file by extension (MIME types taken from https://www.freeformatter.com/mime-types-list.html)"

    parts = filename.split(".")
    if len(parts)<2:
        # The file has no extension.
        # Default to application/octet-stream
        return "application/octet-stream"

    # The extension is whatever is after the last '.' in the filename
    # Switch to lowercase for comparison
    extension = parts[len(parts)-1].lower()

    # Giant dictionary of extensions -> MIME types
    dictionary = {
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

    if not extension in dictionary.keys():
        # We don't recognize this filetype
        # Default to application/octet-stream
        return "application/octet-stream"

    # Recognized filetype. Return it.
    return dictionary[extension]

def basicHeaders(status, contentType):
    "Constructs and returns a basic set of headers for a response (Does not end the header block)"

    out =  "HTTP/1.1 "+status+"\r\n"
    out += "Date: "+time.strftime("%a, %d %b %Y %H:%M:%S GMT", time.gmtime())+"\r\n"
    out += "Server: Cadence purpose-built webserver\r\n"
    out += "Connection: close\r\n"

    # Add cache-control header iff we have caching set
    if caching>0:
        out += "Cache-Control: public, max-age="+str(caching)+"\r\n"

    out += "Content-Type: "+contentType+"\r\n"
    return out

def constructResponse(unendedHeaders, content):
    "Attaches unendedHeaders and content into one HTTP response (adding content-length in the process)"

    response =  unendedHeaders

    # Add ETag iff we have caching set
    if caching>0:
        response += "ETag: \""+hashlib.sha256(content).hexdigest()+"\"\r\n"

    response += "Content-Length: "+str(len(content))+"\r\n\r\n"
    response += content
    return response

def sendResponse(status, contentType, content, sock):
    "Constructs and sends a response with the first three parameters via sock"

    sock.sendall(constructResponse(basicHeaders(status, contentType), content))

# Probably won't see much use for this... But need it at least for 400 bad request
def generateErrorPage(title, description):
    "Returns the HTML for an error page with title and description"

    content =  "<!DOCTYPE html>\n"
    content += "<html>\n"
    content += "  <head>\n"
    content += "    <title>"+title+"</title>\n"
    content += "  </head>\n"
    content += "  <body>\n"
    content += "    <h1 style='text-align: center; width:100%'>"+title+"</h1>\n"
    content += "    <p>"+description+"</p>\n"
    content += "  </body>\n"
    content += "</html>\n"
    return content

# Class to store an open connection
class Connection:
    def __init__(self, conn, isAccept=False):
        self.conn = conn
        self.isAccept = isAccept

    # For compatibility with select
    def fileno(self):
        return self.conn.fileno()

# List of open connections
openconn = []

# Infinite loop for connection service
while True:
    # List of sockets we're waiting to read from
    # (we do block on writes... But we don't want to wait on reads.)
    r = []
    # Add all waiting connections
    for conn in openconn:
        r.append(Connection(conn))
    # And also the incoming connection accept socket
    r.append(Connection(sock, True))

    # Now, select sockets to read from
    readable, u1, u2 = select.select(r, [], [])

    # And process all those sockets
    for read in readable:
        # For the accept socket, accept the connection and add it to the list
        if read.isAccept:
            openconn.append(read.conn.accept()[0])
        else:
            # Fetch the HTTP request waiting on read
            request = waitingRequest(read.conn)
            # Lines of the HTTP request (needed to read the header)
            lines = request.split("\r\n")

            # The first line tells us what we're doing
            # If it's GET, we return the file specified via commandline
            # If it's HEAD, we return the headers we'd return for that file
            # If it's something else, return 405 Method Not Allowed
            method = lines[0]
            if not (method.startswith("GET") or method.startswith("HEAD")):
                # This server can't do anything with these methods.
                # So just tell the browser it's an invalid request
                sendResponse("405 Method Not Allowed",
                             "text/html",
                             generateErrorPage("405 Method Not Allowed",
                                               "Your browser sent a request to perform an action the server doesn't recognize."),
                             read.conn)
                read.conn.close()
                openconn.remove(read.conn)
                continue

            # Parse the filename out of the request
            filename = directory+method.split(' ')[1]
            # If the filename ends in a slash, assume 'index.html'
            if filename.endswith('/'):
                filename += "index.html"

            # Guess the MIME type of the file.
            type = mimeTypeOf(filename)

            # Read the file into memory
            file = ""
            try:
                with open(filename, 'r', 0) as f:
                    file = f.read()
            except FileNotFoundError:
                # The file wasn't found. Return 404.
                sendResponse("404 Not Found",
                             "text/html",
                             generateErrorPage("404 Not Found",
                                               "The requested file \""+method.split(' ')[1]+
                                               "\" was not found on this server."),
                             read.conn)
                # Close the connection and continue
                read.conn.close()
                openconn.remove(read.conn)
                continue
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
                # Close the connection and continue
                read.conn.close()
                openconn.remove(read.conn)
                continue

            # Serve the file back to the client.
            # If the method is GET, use sendResponse to send the file contents.
            if method.startswith("GET"):
                sendResponse("200 OK", type, file, read.conn)
            # If the method is HEAD, generate the same response, but strip the body
            else:
                read.conn.sendall(constructResponse(basicHeaders("200 OK", type), file).split("\r\n\r\n")[0]+"\r\n\r\n")

            # Now that we're done, close the connection and move on.
            read.conn.close()
            openconn.remove(read.conn)
