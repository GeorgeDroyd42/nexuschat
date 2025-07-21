class WebSocketQueue {
    constructor() {
        this.messageQueue = [];
        this.websocket = null;
    }

    setWebSocket(ws) {
        this.websocket = ws;
    }

    sendMessage(data) {
        if (this.websocket && this.websocket.readyState === WebSocket.OPEN) {
            console.log('Sending message via WebSocket:', data);
            this.websocket.send(JSON.stringify(data));
            return true;
        } else {
            console.log('WebSocket not ready, queuing message:', data);
            this.messageQueue.push(data);
            return true;
        }
    }

    flushQueue() {
        if (this.messageQueue.length > 0) {
            console.log('Flushing', this.messageQueue.length, 'queued messages');
            while (this.messageQueue.length > 0) {
                const data = this.messageQueue.shift();
                console.log('Sending queued message:', data);
                this.websocket.send(JSON.stringify(data));
            }
        }
    }

    getQueueSize() {
        return this.messageQueue.length;
    }
}

const wsQueue = new WebSocketQueue();
window.wsQueue = wsQueue;