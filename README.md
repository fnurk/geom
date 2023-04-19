# GEOM
Very cool name. 

Experimental backend to maybe use for quick PoCs in the future. 
Inspired by pocketbase, but targeting a more code-centric usage.

### Why?
Firebase is nice, but self hosted and everything in code is even nicer. 

The goal is to have: 
 - Single file database - easy to back up and handle
   - Schemaless documents for quick development and wild west migrations
   - If anyone asks "where is our data?" the answer is "this file"
   - Easy to separate data into files, eg GDPR data in one, anonymous data in another, a third for time-series data
 - Easy to implement access control
   - Firebase-esque rules for documents
 - Tools ready to implement business logic
   - In memory pubsub
   - In memory cache
   - Realtime document updates
 - Indexes from json document fields
   - Simplify "joins"
 - Acceptable performance for a few thousand users
   - Clear paths forward when performance starts degrading
     - Easy to replace in embedded pubsub/cache/dbs with externals
 - Super simple deployment
   - Exetutable + database file + config file -> rock and roll

