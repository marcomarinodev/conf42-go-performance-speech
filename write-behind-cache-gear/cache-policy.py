# Gears Recipe for a single write behind

# import redis gears & mongo db libs
from rgsync import RGJSONWriteBehind
from rgsync.Connectors import MongoConnection, MongoConnector
import json

# change mongodb connection
# MongoConnection(user, password, host, authSource (optional), fullConnectionUrl (optional) )
# connection = MongoConnection('ADMIN_USER','ADMIN_PASSWORD','ADMIN_HOST', "admin")
connection = MongoConnection("", "", "", "", "mongodb://mongodb_server:27017/store")

# change MongoDB database
db = 'store'

# change MongoDB collection & it's primary key
userConnector = MongoConnector(connection, db, 'users', 'userID')

def fetch_data(r):
    key = r['key']
    log('Key %s was fetched and missed' % key)
    try:
        # Debugging find_one function with prints
        log('Fetching data for key: %s' % key)
        collection = userConnector.collection  # Retrieve the MongoDB collection object
        data = collection.find_one({'userID': key})
        if data:
            log("Data fetched from MongoDB for key %s: %s" % (key, data))
            # Convert data to JSON and ignore _id property
            json_data = json.dumps(data, default=str)
            # Execute the set command with JSON data
            execute('json.set', key, json_data)
        else:
            log("No data found for key %s in MongoDB" % key)
    except Exception as e:
        log("Error fetching data for key %s: %s" % (key, str(e)))

GB().map(fetch_data).register(eventTypes=['keymiss'])

# change redis keys with prefix that must be synced with mongodb collection
RGJSONWriteBehind(GB, keysPrefix='UserEntity',
                  connector=userConnector, name='UsersWriteBehind',
                  version='99.99.99')
