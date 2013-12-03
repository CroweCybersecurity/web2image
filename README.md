web2image
=========

PhantomJS tool to convert web pages to screenshots.  This requires the PhantomJS project (http://www.phantomjs.org or https://github.com/ariya/phantomjs/).  Once you have the phantomjs.exe, simply run the JavaScript file using the following arguments: 

```
Usage: phantomjs.exe --ignore-ssl-errors=yes web2image.js <URL LIST> <OUTFILE>
```
- URL LIST - File with one URL per line (http://domain.com:port/)
- OUTFILE  - An HTML file to allow for pretty output

To use a proxy: phantomjs.exe --proxy=address:port web2image.js
