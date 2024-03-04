//--web true
//--docker ghcr.io/nuvolaris/runtime-nodejs-v21:3.1.0-mastrogpt.2403032235
//--param OPENAI_API_KEY $OPENAI_API_KEY
//--param OPENAI_API_HOST $OPENAI_API_HOST

const { OpenAIClient, AzureKeyCredential } = require("@azure/openai");

async function main(args) {
    const key = args.OPENAI_API_KEY || process.env.OPENAI_API_KEY
    const host = args.OPENAI_API_HOST || process.env.OPENAI_API_HOST
    const model = "gpt-35-turbo"
    const AI = new OpenAIClient(host, new AzureKeyCredential(key))
    const input = args.input || ""
    let answer = "Please provide an input parameter."
    if (input != "") {
        //const { id, created, choices, usage } =
        const request = [
            { role: "system", content: "You are a helpful assistant." },
            { role: "user", content: input },
        ]
        const response = await AI.getChatCompletions(model, request);
        answer = response.choices[0].message.content
    }
    return { body: answer }
}