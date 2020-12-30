package main

//In this file all parameters for all clients are set

//GENERAL PARAMETERS
const defaultGameURL = "ws://localhost:8080/spe_ed"

const defaultTimeURL = "http://localhost:8080/spe_ed_time"

//Name of logging file
const loggingFile = "logging.txt"

//This value is specified in milliseconds and is a reserve in which the actions should be send to the server
const calculationTimeOffset = 150

//If getting the time from the timing api fails and the calculation time calculated is more than this value in minutes away the program will throw an error
const maxCalculationTime = 2

//This value specifies under which address you reach the webinterface
const guiURL = "0.0.0.0:8081"

//This defines the max speed a player can reach in our simulation. To simulate most realistic set it to 10
const maxSpeed = 6

//MINIMAX PARAMETERS

//If minimax evaluates multiple actions equally good he will use this order to choose which one to use
var minimaxPreferences []Action = []Action{ChangeNothing, TurnLeft, TurnRight, SpeedUp, SlowDown}

//PROBABILITY TABLE PARAMETERS

//This const defines the maximal number of Turns simulateGame will try to process
const maxSimDepth = 20

//This const defines after how many processed players simulatePlayer will schedule a garbage Collection cycle. Lowering the value improves memory efficiency but has a performance impact
const processedPlayersTillGC = 60000

//ROLLOUT PARAMETERS

//If this value is set to true we process in every rollout before we choose our own action a action for every other living player
const simulateOtherPlayers = false

//This const defines the max number of Rollouts simulateRollouts will perform. Normally there is no good reason to change this value
const maxNumberofRollouts = 7000000

//This const defines the relation between the longest and the shortest path simulateRollouts gives back
const filterValue = 0.75
