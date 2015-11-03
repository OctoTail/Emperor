#!/usr/bin/python
from time import sleep
import socketserver
import threading
import logging
import sys
import hmac
import base64

class UDPHandler(socketserver.BaseRequestHandler):
    """
    Request structure:
    [player_id]\n[HMAC(md5) in base64 minus the final '\n']\0[request]
    """
    def handle(self):
        global dataLock
        global players
        socket = self.request[1]
        (auth,s,msg) = self.request[0].partition(b'\0')
        dataLock.acquire()
        try:
            if(len(s)==0):
                raise ValueError("Bad message (no HMAC?)")
            (plId,digest)=auth.split(b'\n')
            try:
                pl=players[plId.decode()]
            except:
                raise ValueError("Invalid user")
            if(digest!=base64.encodestring(hmac.new(pl["key"],msg).digest())[:-1]):
                raise ValueError("Wrong HMAC")
            if(len(msg)==0): #Heartbeat
                socket.sendto(getPlData(pl),self.client_address)
        except Exception as e:
            logging.warning(e)
        finally:
            dataLock.release()

class ObjectDict(dict):
    def __init__(self,_set,field):
        return super().__init__({i[field]:i for i in _set})

    def __init__(self,*args,**kwargs):
        return super().__init__({i[kwargs["field"]]:i for i in args})

    def __iter__(self):
        return iter(super().values())

class Wrapper():
    def __init__(self,_dict):
        self._dict=_dict

    def __init__(self,**kwargs):
        self._dict=kwargs

    def __getitem__(self,key):
        return self._dict.__getitem__(key)

    def __setitem__(self,key,value):
        return self._dict.__setitem__(key)

class City(Wrapper):
    pass

class Player(Wrapper):
    pass

class Unit(Wrapper):
    pass

def mainF():
    """
    Main loop, updates every 100ms
    While running, locks the data
    """
    global dataLock
    global exitNow
    while(not exitNow):
        dataLock.acquire()
        dataLock.release()
        sleep(.1)

def getPlData(pl):
    """
    Data structure:
    [city1]\n[city2]\0[pl1]\n[pl2]
    """
    cityData=[]
    for city in cities:
        if(city["owner"]==pl["name"]):
            cityData.append(packData(city._dict)) #Return full data
    cityData=b'\n'.join(cityData)
    plData=[]
    for player in players:
        if(player is pl):
            plData.append(packData(dict(filter((lambda x:x[0]!="key"),pl._dict.items()))))
    plData=b'\n'.join(plData)
    return b'\0'.join((cityData,plData)) #Return full data

def packData(dict_data):
    """
    Used to pack dict data to be sent to the client
    For dict_data={key1:value1,key2:[value21,value22],key3:{key31,key32}, ...}
    return b'key1:value1;key2:value21,value22;key3:key31,key32'
    Values and keys are first converted to str
    """
    list_data=[]
    for (key,value) in dict_data.items():
        if(type(value) is set or type(value) is list or type(value) is tuple):
            list_data.append(str(key)+':'+','.join((str(i) for i in value)))
        else:
            list_data.append(str(key)+':'+str(value))
    return ';'.join(list_data).encode()

def initData():
    global players
    global cities
    cities=ObjectDict(City(name="Paradizna",owner="root",pop=10000,loc=(10,15)),
            City(name="True Glorian",owner="isaac",pop=500,loc=(5,5)),
            field="name")
    players=ObjectDict(Player(name="root",key=b"supersecret",gold=1000000),
            Player(name="isaac",key=b"abaddon",gold=5000),
            field="name")
    units=ObjectDict(Unit(name=0,content={"infantry":100},owner="root"),
            Unit(name=1,content={"archer":50},owner="isaac",position=(5,6)),
            field="name")

if(__name__ == "__main__"):
    global exitNow
    global dataLock
    logging.root.setLevel('DEBUG') #debug
    HOST=""
    PORT=int(sys.argv[1])
    try:
        UDPserver = socketserver.UDPServer((HOST,PORT),UDPHandler)
    except:
        logging.error("Failed to create a UDP socketserver!")
        exit(1)
    initData()
    dataLock=threading.Lock()
    exitNow=False
    mainThread=threading.Thread(target=mainF)
    mainThread.start()
    try:
        UDPserver.serve_forever()
    except KeyboardInterrupt:
        logging.error("Keyboard interrupt!")
        exitNow=True

