# FAQ

> Why do I get a payment failure with NO_ROUTE ?  

This is due to your payment lightning node not finding a route to the merchant lightning node.  
It may help to open well funded channels to other well connected nodes.  
It may also help to increase your maximum network fee in your service variables, **FEE_LIMIT_SAT** .  
It can be useful to test paying invoices directly from your lightning node.  

> Why do my payments take so long ?  


This is due to the time taken for your payment lightning node to find a route.  
It can be improved by opening channels using clearnet rather than on the tor network.  
It may also help to improve your lightning node hardware or software setup.  
It can be useful to test paying invoices directly from your lightning node.  
