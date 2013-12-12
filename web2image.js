var fs = require('fs'),
	system = require('system'),
	crlf = system.os.name == 'windows' ? "\r\n" : "\n",
	outfile = null,
	hostfile = '';

function usage() {
	console.log("Usage: phantomjs.exe --ignore-ssl-errors=yes web2image.js <URL LIST> <OUTFILE>");
	console.log("  URL LIST - File with one URL per line (http://domain.com:port/)");
	console.log("  OUTFILE  - An HTML file to allow for pretty output");
	console.log("  To use a proxy: phantomjs.exe --proxy=address:port web2image.js");
}

function makefilename(url) {
	var filename = url.split("://");
	filename = filename[0] + "_" + filename[1] + ".png";
	filename = filename.replace("/", "");
	filename = filename.replace(":", "_");
	return filename;
}

var render = function(urls, callbackPerUrl, callbackFinal) {
	var next, page, retrieve, webpage;
	webpage = require("webpage");
	page = null;

	next = function(status, url, file, response, error) {
		page.close();
		callbackPerUrl(status, url, file, response, error);
		return retrieve();
	};

	retrieve = function() {
		var url, response, error;
		if (urls.length > 0) {
			url = urls.shift();

			page = webpage.create();
			page.viewportSize = {
				width: 800,
				height: 600
			};

			page.settings.userAgent = "Web2Image";
			page.settings.webSecurityEnabled = false;
			page.settings.resourceTimeout = 10000;
			page.onResourceReceived = function(res) {
				if(res.url === page.url) {
					response = res;
				}
			};
			page.onResourceError = function(res) {
				if(res.url === page.url) {
					error = res;
				}
			};
			page.onResourceTimeout = function(res) {
				if(res.url === page.url) {
					error = res;
				}
			};

			return page.open(url, function(status) {
			var filename = makefilename(url);
			if (status === "success") {
				return window.setTimeout((function() {
					page.render(filename);
					return next(status, url, filename, response, error);
				}), 200);
			} else {
				return next(status, url, filename, response, error);
				}
			});
		} else {
			return callbackFinal();
		}
	};
	return retrieve();
};

console.log("Web2Image v1.2");
if (system.args.length != 3) {
	usage();
	phantom.exit(1);
}

try {
	var infile = fs.open(system.args[1], "r");
	hostfile = infile.read();
	infile.close();
} catch (e) {
	usage();
	console.log("ERROR: Unable to open input file.");
	console.log(e);
	phantom.exit(1);
}

try {
	outfile = fs.open(system.args[2], "w");
	var tmp = "<!doctype html> \n\
<html>\n\
<head>\n\
<title>Web2Image</title>\n\
<style type='text/css'>\n\
	#main {\n\
		border: 1px solid #000000;\n\
		border-collapse: collapse;\n\
		font: .8em arial,sans-serif;\n\
	}\n\
	#main th {\n\
		background-color: #005C8A;\n\
		color: #ffffff;\n\
	}\n\
	#main tr.error td {\n\
		background-color: #FFB2B2;\n\
	}\n\
	#main td, #main th {\n\
		padding: 5px;\n\
		vertical-align: top;\n\
		text-align: center;\n\
	}\n\
	#main td.img div {\n\
		width: 100%;\n\
		max-height: 450px;\n\
		overflow: auto;\n\
	}\n\
	#main td img {\n\
		width: 95%;\n\
	}\n\
	#main td.url, #main th.url {\n\
		width: 10%;\n\
		word-break: break-all;\n\
	}\n\
	#main td.img, #main th.img {\n\
		width: 40%;\n\
	}\n\
	#main td.status, #main th.status {\n\
		width: 8%;\n\
	}\n\
	#main td.headers, #main th.headers {\n\
		width: 42%;\n\
	}\n\
	#main table.header {\n\
		width: 100%;\n\
		word-break: break-all;\n\
	}\n\
	#main table.header td.name {\n\
		width: 25%;\n\
	}\n\
	#main table.header td {\n\
		border-top: 1px dotted #444444;\n\
		border-collapse: collapse;\n\
		text-align: left;\n\
	}\n\
	table.header {\n\
		border-bottom: 1px dotted #444444;\n\
	}\n\
</style>\n\
</head>\n\
<body>\n\
<table id='main' border=1>\n\
	<thead>\n\
	<tr>\n\
		<th class='url'>URL</th>\n\
		<th class='img'>Image</th>\n\
		<th class='status'>Response Code</th>\n\
		<th class='headers'>Headers</th>\n\
	</tr>\n\
	</thead>\n\
	<tbody>";
	outfile.writeLine(tmp);
} catch (e) {
	usage();
	console.log("ERROR: Unable to open output file.");
	console.log(e);
	phantom.exit(1);
}

var hosts = hostfile.split(crlf);
render(hosts, (function(status, url, file, response, error) {
	if(status !== "success") {
		var message, code;

		message = error !== undefined && error.errorString !== undefined ? error.errorString : "Error was undefined.";
		code = error !== undefined && error.errorCode !== undefined ? error.errorCode : "&#9785;";

		outfile.writeLine("\t<tr class='error'>");
		outfile.writeLine("\t\t<td class='url'><a target='_blank' href='"+url+"'>"+url+"</a></td>");
		outfile.writeLine("\t\t<td class='img' colspan=3>An error occured ("+code+") while loading the page: "+message+"</td>");
		outfile.writeLine("\t</tr>");
		return console.log("[!] ERROR: Unable to render '" + url + "'");
	} else {
		try {
			message = response !== undefined && response.statusText !== undefined ? response.statusText : "Unknown";
			code = response !== undefined && response.status !== undefined ? response.status : "000";

			if(response !== undefined && response.headers !== undefined) {
				var headerstring = "<table class='header'>";
				for(var i = 0; i < response.headers.length; i++) {
					headerstring += "<tr>";
					headerstring += "<td class='name'>"+response.headers[i]['name']+"</td>";
					headerstring += "<td class='value'>"+response.headers[i]['value']+"</td>";
					headerstring += "</tr>";
				}
				headerstring += "</table>";
			} else {
				headerstring = "Unable to gather header information."
			}
			outfile.writeLine("\t<tr>");
			outfile.writeLine("\t\t<td class='url'><a target='_blank' href='"+url+"'>"+url+"</a></td>");
			outfile.writeLine("\t\t<td class='img'><div><a target='_blank' href='"+makefilename(url)+"'><img src='"+makefilename(url)+"'/></a></div></td>");
			outfile.writeLine("\t\t<td class='status'>"+code+" "+message+"</td>");
			outfile.writeLine("\t\t<td class='headers'>"+headerstring+"</td>");
			outfile.writeLine("\t</tr>");

			outfile.flush();

			return console.log("[*] Rendered '" + url + "' to '" + file + "'");
		} catch (e) {
			usage();
			console.log("ERROR: Unable to write to output file.");
			console.log(e);
		}
	}
}), function() {
	try {
		outfile.writeLine("\t</tbody>");
		outfile.writeLine("</table>");
		outfile.writeLine("</body>");
		outfile.writeLine("</html>");
		outfile.close();
	} catch (e) {
		console.log("ERROR: Something bad happened when closing the output file.");
		console.log(e);
	}
	return phantom.exit();
});
