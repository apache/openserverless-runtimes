function hello(args) {
    let name = args['name'] || "world" 
    return `Hello, ${name}.`
}

module.exports = hello