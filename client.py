import asyncio
import websockets
import json

#JSON to send for corresponding action
LEFT = '{"action":"turn_left"}'
RIGHT = '{"action":"turn_right"}'
SLOWER = '{"action":"slow_down"}'
FASTER = '{"action":"speed_up"}'
NOTHING = '{"action":"change_nothing"}' 


async def connect():
    uri = "ws://localhost:8080/spe_ed"
    #conect to websocket
    async with websockets.connect(uri) as websocket:
        while(True):
            # wait for data from server
            rxData = await websocket.recv()
            # use data to calculate next move
            msg = useData(rxData)
            # send next move back to server
            await websocket.send(msg)

# useData uses the received game data to return the next actions
def useData(data):
    # load json data
    d = json.loads(data)
    # extract information from json (examples)
    deadline = d["deadline"]
    you = d["you"]
    own_direction = d["players"][str(you)]["direction"]
    #...

    #print("Your dir:" + d["players"][str(you)]["direction"])
    #print("Other dir:" + d["players"][str(you+1)]["direction"])

    #move diagonally
    if(own_direction == "right"):
        return RIGHT
    elif(own_direction == "down"):
        return LEFT
    
    return NOTHING



#main
asyncio.get_event_loop().run_until_complete(connect())
asyncio.get_event_loop().run_forever()
print("Fertig")