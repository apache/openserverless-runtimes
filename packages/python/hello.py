#--web true
#--kind python:default
def main(args):
    name = args.get("name", "world")
    return {
        "body": f"Hello, {name}."
    }





