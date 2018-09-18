make && 
serverless deploy  
echo "-------------- Query Calc --------------"
serverless invoke -f calc
echo "-------------- Query Logs --------------"
serverless invoke -f logs --path test-events/roamerRequestIds.json
echo "-------------- Query Orders --------------"
serverless invoke -f orders --path test-events/startAndEnd.json
echo "-------------- Query Events --------------"
serverless invoke -f events --path test-events/startAndEndAndOrders.json