package main

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
