const tracer = require('dd-trace').init({ logInjection: true });
const express = require('express');
const helmet = require('helmet');

const app = express();
app.use(helmet())

app.get('/', tracer.wrap('handler', (req, res) => {
    res.send('Hello world from node.js!');
}));

const port = process.env.PORT || 8080;

app.listen(port, () => {
    console.log(`Server listening on port ${port}`);
});