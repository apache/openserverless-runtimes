//--web true
//--docker ghcr.io/nuvolaris/runtime-nodejs-v21:3.1.0-mastrogpt.2402201748


const marked = require("marked");

function main(args) {
    let name = args.name || "world".
    let text = `# Welcome\n\nHello, ${name}.`
    return {
        body:  marked.parse(text)
    }
}

module.exports = main