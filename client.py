import asyncio
import websockets
import json

# JSON to send for corresponding action
LEFT = '{"action":"turn_left"}'
RIGHT = '{"action":"turn_right"}'
SLOWER = '{"action":"slow_down"}'
FASTER = '{"action":"speed_up"}'
NOTHING = '{"action":"change_nothing"}' 

# directions
NORTH = "up"
EAST = "right"
SOUTH = "down"
WEST = "left"

# more variables
uri = "ws://localhost:8080/spe_ed"

class info(object):
    """Creates an info object with all the information sent by the server"""
    def __init__(self, data):
        d = json.loads(data)
        self.width = d["width"]
        self.height = d["height"]
        self.cells = d["cells"]
        self.players = d["players"]
        self.you = d["you"]
        self.running = d["running"]
        self.deadline = d["deadline"]

        self.player_you = player(self.players[str(self.you)])



class player(object):
    """Creates a player object from the given json"""
    def __init__(self, data):
        self.x = data["x"]
        self.y = data["y"]
        self.direction = data["direction"]
        self.speed = data["speed"]
        self.active = data["active"]
        self.name = data["name"]



async def connect():
    #conect to websocket
    async with websockets.connect(uri) as websocket:
        while(True):
            # wait for data from server
            rxData = await websocket.recv()

            # convert received json to info object
            information = info(rxData)

            # calculate move from information
            move = useData(information)

            # send next move back to server
            await websocket.send(move)

def fieldView(viewDist, information):
    """Uses the viewing distance and player object to return a field of view"""
    cells = []
    start_x = information.player_you.x - viewDist
    start_y = information.player_you.y - viewDist
    for i in range(viewDist*2+1):
        row =[]
        for j in range(viewDist*2+1):
            index_x = start_x + j  
            index_y = start_y + i  
            if (index_x >= 0 and index_y >= 0 and index_x < information.width and index_y < information.height):
                data = information.cells[index_x][index_y]
                row.append(data)
            else:
                row.append(-1)
        cells.append(row)
    return cells

def useData(information):
    # here you can put code to calculate the next move from given data
    fieldView(4,information)


    # return the move
    return NOTHING




#main
asyncio.get_event_loop().run_until_complete(connect())
asyncio.get_event_loop().run_forever()
print("Fertig")