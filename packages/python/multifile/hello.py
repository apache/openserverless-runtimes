def hello(args):
    name = args.get("name", "world")
    return "Hello, f{name}."