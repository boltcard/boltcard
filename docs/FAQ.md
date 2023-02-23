# How do you bech32 encode a string on the card ?

The LNURLw that comes from the bolt card is not bech32 encoded.
It uses [LUD-17](https://github.com/fiatjaf/lnurl-rfc/blob/luds/17.md).

# How do I generate a random key value ?

This will give you a new 128 bit random key as a 32 character hex value.  
`$ hexdump -vn16 -e'4/4 "%08x" 1 "\n"' /dev/random`

# Why do I get a payment failure with NO_ROUTE ?  

This is due to your payment lightning node not finding a route to the merchant lightning node.  
It may help to open well funded channels to other well connected nodes.  
It may also help to increase your maximum network fee in your service variables, **FEE_LIMIT_SAT** / **FEE_LIMIT_PERCENT** .  
It can be useful to test paying invoices directly from your lightning node.  

# Why do my payments take so long ?  

This is due to the time taken for your payment lightning node to find a route.  
It can be improved by opening channels using clearnet rather than on the tor network.  
It may also help to improve your lightning node hardware or software setup.  
It can be useful to test paying invoices directly from your lightning node.  

# Can I use the same lightning node for the customer (bolt card) and the merchant (POS) ?

When tested with LND in Nov 2022, the paying (customer, bolt card) lightning node must be a separate instance to the invoicing (merchant, POS) lightning node.
