/* eslint-disable no-console */
const http = require('http');

const book = {
  urn: 'relativity-the-special-general-theory'
};

http
  .createServer(function (req, res) {
    console.log('STARTED LIBRARY APPLICATION');
    if (req.url && req.url === '/books/relativity-the-special-general-theory') {
      res.writeHead(200, { 'Content-Type': 'application.json' });
      res.write(JSON.stringify({ ...book }));
      return res.end();
    }

    res.writeHead(404, { 'Content-Type': 'application.json' });
    return res.end(JSON.stringify({ message: 'book not found' }));
  })
  .listen(4000);
