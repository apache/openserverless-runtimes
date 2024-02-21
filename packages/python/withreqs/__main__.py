#--web true
#--kind python:default

import htmlgenerator as hg

def main(args):
    name = args.get("name", "world")
    page = hg.HTML(hg.HEAD(), hg.BODY(hg.H1(f"Hello, {name}.")))
    return {
        "body": hg.render(page, {})
    }