const ROOMS = ["general", "games"];

const $ = (id) => document.getElementById(id);

const loginForm = $("login");
const loginName = $("loginName");
const loginPassword = $("loginPassword");
const roomSelect = $("roomSelect");
const loginError = $("loginError");

const chat = $("chat");
const list = $("messages");
const statusDot = $("status");
const roomTitle = $("roomName");
const onlineCount = $("onlineCount");
const msgInput = $("msgInput");

let ws = null;
let myName = "";
let lastError = "";
let unread = 0;

// ---------- комната из адреса: /home?room=games ----------

function roomFromURL() {
    const room = new URLSearchParams(location.search).get("room");
    return ROOMS.includes(room) ? room : "general";
}

roomSelect.value = roomFromURL();

roomSelect.addEventListener("change", () => {
    history.replaceState(null, "", "?room=" + encodeURIComponent(roomSelect.value));
});

// ---------- регистрация и вход ----------

async function postCredentials(url) {
    const body = new URLSearchParams({
        name: loginName.value.trim(),
        password: loginPassword.value,
    });
    return fetch(url, { method: "POST", body });
}

async function register() {
    loginError.textContent = "";
    const res = await postCredentials("/register");
    if (res.status === 201) {
        await login(); // регистрация сразу входит
        return;
    }
    loginError.textContent = await res.text();
}

async function login() {
    loginError.textContent = "";
    const res = await postCredentials("/login");
    if (!res.ok) {
        loginError.textContent = await res.text();
        return;
    }
    myName = loginName.value.trim();
    loginPassword.value = "";
    connect(roomSelect.value);
}

// ---------- экраны ----------

function showLogin(message) {
    ws = null;
    chat.hidden = true;
    loginForm.hidden = false;
    loginError.textContent = message;
}

// ---------- WebSocket ----------

function connect(room) {
    const proto = location.protocol === "https:" ? "wss://" : "ws://";

    list.innerHTML = "";
    lastError = "";
    roomTitle.textContent = "#" + room;

    // имени в адресе нет: сервер берёт его из сессии по cookie
    ws = new WebSocket(proto + location.host + "/ws?room=" + encodeURIComponent(room));

    ws.onopen = () => {
        statusDot.classList.add("online");
        loginForm.hidden = true;
        chat.hidden = false;
        msgInput.focus();
    };

    ws.onmessage = (e) => handleMessage(e.data);

    ws.onclose = () => {
        statusDot.classList.remove("online");
        // если cookie нет, сервер ответит 401 и onopen не сработает
        showLogin(lastError || "Соединение закрыто. Войди снова.");
    };
}

function handleMessage(data) {
    if (data.startsWith("ОШИБКА: ")) {
        lastError = data.slice("ОШИБКА: ".length);
        addSystem(lastError);
        return;
    }

    if (data.startsWith("CMD:online:")) {
        const raw = data.slice("CMD:online:".length);
        const names = raw === "" ? [] : raw.split(",");
        onlineCount.textContent = "Онлайн: " + names.length;
        onlineCount.title = names.join(", ");
        return;
    }

    if (data.startsWith("CMD:user_joined:")) {
        addSystem("Пользователь " + data.slice("CMD:user_joined:".length) + " присоединился");
    } else if (data.startsWith("CMD:user_left:")) {
        addSystem("Пользователь " + data.slice("CMD:user_left:".length) + " покинул чат");
    } else if (data === "CMD:show_confirm") {
        if (confirm("Удалить весь чат?")) {
            ws.send("CMD:confirm_yes");
        }
    } else if (data === "CMD:cleared") {
        list.innerHTML = "";
        addSystem("Chat cleared");
    } else if (!data.startsWith("CMD:")) {
        addMessage(data);
    }
}

// ---------- отображение ----------

function nickColor(nick) {
    let h = 0;
    for (const ch of nick) {
        h = (h * 31 + ch.charCodeAt(0)) % 360;
    }
    return "hsl(" + h + ", 70%, 72%)";
}

function addSystem(text) {
    const li = document.createElement("li");
    li.classList.add("system");
    li.textContent = text;
    list.appendChild(li);
    list.scrollTop = list.scrollHeight;
}

function addMessage(data) {
    // сервер присылает "ник | текст"
    const sep = data.indexOf(" | ");
    const nick = sep === -1 ? "" : data.slice(0, sep);
    const text = sep === -1 ? data : data.slice(sep + 3);
    const mine = nick === myName;

    const li = document.createElement("li");
    if (mine) li.classList.add("mine");

    if (!mine && nick) {
        const who = document.createElement("span");
        who.className = "who";
        who.textContent = nick;
        who.style.color = nickColor(nick);
        li.appendChild(who);
    }

    const body = document.createElement("span");
    body.textContent = text;
    li.appendChild(body);

    const time = document.createElement("span");
    time.className = "time";
    time.textContent = new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
    li.appendChild(time);

    list.appendChild(li);
    list.scrollTop = list.scrollHeight;

    if (document.hidden && !mine) {
        unread++;
        document.title = "(" + unread + ") Chat";
    }
}

document.addEventListener("visibilitychange", () => {
    if (!document.hidden) {
        unread = 0;
        document.title = "Chat";
    }
});

// ---------- события ----------

$("loginBtn").addEventListener("click", login);
$("registerBtn").addEventListener("click", register);

loginPassword.addEventListener("keydown", (e) => {
    if (e.key === "Enter") login();
});

$("send-msg").addEventListener("submit", (e) => {
    e.preventDefault();
    const text = msgInput.value.trim();
    if (text === "" || !ws || ws.readyState !== WebSocket.OPEN) return;
    ws.send(text);
    msgInput.value = "";
});

$("deleteBtn").addEventListener("click", () => {
    if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send("CMD:show_confirm");
    }
});