package main

//In this file all parameters for all clients are set

//This value is specified in milliseconds and is a reserve in which the actions should be send to the server
const calculationTimeOffset = 150

//If getting the time from the timing api fails and the calculation time calculated is more than this value in minutes away the program will throw an error
const maxCalculationTime = 2

//This defines the max speed a player can reach in our simulation. To simulate most realistic set it to 10
const maxSpeed = 6

//MINIMAX PARAMETERS

//PROBABILITY TABLE PARAMETERS

//Setting this value higher, makes makes us more confident, that we will visit a field and no one other does. Setting it lower does the exact opposite
const myStartProbability = 1.0

const othersStartProbability = 1.0

//This const defines the maximal number of Turns simulateGame will try to process
const maxSimDepth = 40

//This const defines after how many processed players simulatePlayer will schedule a garbage Collection cycle. Lowering the value improves memory efficiency but has a performance impact
const processedPlayersTillGC = 60000

//ROLLOUT PARAMETERS

//If this value is set to true we process in every rollout before we choose our own action a action for every other living player
const simulateOtherPlayers = false

//This const defines the max number of Rollouts simulateRollouts will perform. Normally there is no good reason to change this value
const maxNumberofRollouts = 7000000

//This const defines the relation between the longest and the shortest path simulateRollouts gives back
const filterValue = 0.75
