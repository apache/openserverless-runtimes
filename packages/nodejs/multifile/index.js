//--web true
//--docker ghcr.io/nuvolaris/runtime-nodejs-v21:3.1.0-mastrogpt.2402201748


const hello = require("./hello")

function main(args) { 
    return { 
        body: hello()
    }
}

module.exports = main