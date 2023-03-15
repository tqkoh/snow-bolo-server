# websocket protocol

`wss://tqk.trap.show/snow-bolo-server/api/ws`

## summary

数 F に一回、サーバーから update が送られる。  
join を送ると joinAccepted が送られ、ゲームに参加できる。input で入力を送るが、送らなくてもサーバーで前の F の入力が使われるため、入力が変更されたときだけ送っている。  
死ぬと dead が送られるが、コネクションは切断されず、update が送られ続ける。(クライアントは spectator モードができる) もう一回 join を送るとまた参加できる。  

json 形式で送る。

```json
{
  "method": "string",
  "args": {
    "key": "object",
    // ... 
  }
}
```

のような形式

## サーバーへの送信

### join

ゲームに参加するリクエストを送る

```json
{
  "method": "json",
  "args": {
    "name": "string"
  }
}
```

### input

入力を送る

- dy, dx は、マウスの中心に対しての相対座標

```json
{
  "method": "input",
  "args": {
    "w": true,
    "a": false,
    "s": false,
    "d": false,
    "left": false,
    "right": false,
    "dy": 0,
    "dx": 0
  }
}
```

### active

アクティブになったとき、インアクティブになったときに送る
active: false を送るとサーバー側からの update 送信が止まり、クライアントが重くなるのを防ぐ

```json
{
  "method": "active",
  "args": {
    "active": true
  }
}
```

## サーバーからの受信

### joinAccepted

参加できたときに送られる

- id: プレイヤーの id

```json
{
  "method": "joinAccepted",
  "args": {
    "id": "00000000-0000-0000-0000-000000000000"
  }
}
```

### update

ゲームの状態が送られる

- users: プレイヤーの情報
  - dummy: ダミープレイヤーかどうか。ダミーはとどめのダメージを表示するために使っていて、クライアントでは描画しない
  - mass, strength: デカさと固さ。クライアントでは切り下げで表示
  - dy, dx: マウスの中心に対しての相対座標
- bullets: 弾の情報
  - owner: 弾を撃ったプレイヤーの id
  - life: 弾の残りのライフ(F)
- feeds: 雪玉の情報

```json
{
  "method": "update",
  "args": {
    "users": [
      {
        "id": "00000000-0000-0000-0000-000000000000",
        "dummy": false,
        "name": "string",
        "mass": 5.0,
        "strength": 100,
        "damage": 10,
        "y": 0.0,
        "x": 0.0,
        "vy": 0.0,
        "vx": 0.0,
        "dy": 0,
        "dx": 0,
        "leftClickLength": 0,
        "rightClickLength": 0,
      },
      // ...
    ],
    "bullets": [
      {
        "id": "00000000-0000-0000-0000-000000000000",
        "owner": "00000000-0000-0000-0000-000000000000",
        "mass": 5.0,
        "life": 60,
        "y": 0.0,
        "x": 0.0,
        "vy": 0.0,
        "vx": 0.0,
      },
      // ...
    ],
    "feeds": [
      {
        "id": "00000000-0000-0000-0000-000000000000",
        "mass": 5.0,
        "y": 0.0,
        "x": 0.0,
        "vy": 0.0,
        "vx": 0.0,
      },
      // ...
    ]
  }
}
```

### message

チャットが送られる
今あるのは

- プレイヤーの参加、切断
- キル

```json
{
  "method": "message",
  "args": {
    "message": "string"
  }
}
```

### dead

死んだときに送られる

```json
{
  "method": "dead",
  "args": {
    "kills": 0,
  }
}
```
