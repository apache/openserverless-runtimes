//--web true
//--docker ghcr.io/nuvolaris/runtime-nodejs-v21:3.1.0-mastrogpt.2402201748

function main(args) {
    let name = args.name || "world"
    return { body: `Hello, ${name}` }
}

