import './index.css'

let t: Talk | null = null;

document.getElementById("username")?.addEventListener("keyup", (event) => {
  if (event.key === "Enter") { connect() }
})

document.getElementById("talk")?.addEventListener("click", () => { connect() })

document.getElementById("dst")?.addEventListener("keyup", (event) => {
  if (event.key === "Enter") { send() }
})
document.getElementById("message")?.addEventListener("keyup", (event) => {
  if (event.key === "Enter") { send() }
})

function connect() {
  const username = (document.getElementById("username") as HTMLInputElement).value
  document.getElementById("talkia")?.classList.remove("hidden")
  document.getElementById("talkia")?.classList.add("flex")
  document.getElementById("connection")?.classList.add("hidden")
  t = Talk("localhost:8080", username)
  t.onMessage((msg) => { createMessage(msg) })
}

function createMessage(msg: Message) {
  const template = document.getElementById('talk-message') as HTMLTemplateElement
  const content = template.content.cloneNode(true) as DocumentFragment
  const talkbox = document.getElementById("talkbox")
  let src = content.querySelector("[slot=src]") as HTMLSpanElement
  if (src !== null) src.innerText = msg.src
  let dst = content.querySelector("[slot=value]") as HTMLSpanElement
  if (dst !== null) dst.innerText = msg.value

  talkbox?.appendChild(content)
  talkbox?.scrollTo({ left: 0, top: talkbox?.scrollHeight, behavior: "smooth" });
}

function send() {
  const dst = (document.getElementById("dst") as HTMLInputElement).value;
  const message = document.getElementById("message") as HTMLInputElement;
  const value = message.value;
  if (value === "") { return; }
  if (dst === "") { return; }
  message.value = "";

  if (t === null) {
    console.error("Not connected");
    return;
  }
  t.send({ tid: "0", dst, src: t.username, status: 0, value });
}

export interface Message {
  tid: string
  dst: string
  src: string
  status: number
  value: string
}

export interface Talk {
  username: string
  send: (msg: Message) => void
  onMessage: (cb: (msg: Message) => void) => void
}

function Talk(host: string, username: string): Talk {
  const url = (location.protocol === "https:" ? "wss" : "ws") + "://" + host + "/ws/" + username
  let t = new WebSocket(url)
  t.onopen = () => { console.log("Connected to " + url) }
  t.onclose = () => { console.log("Disconnected from " + url) }
  t.onerror = (e) => { console.error("Error: " + e) }

  return {
    get username() { return username },
    send: (msg: Message) => {
      msg.src = username
      t.send(JSON.stringify(msg))
    },
    onMessage: (cb: (msg: Message) => void) => {
      t.onmessage = (event) => {
        try { cb(JSON.parse(event.data)) } catch (e) { console.error(e) }
      }
    }
  }
}

