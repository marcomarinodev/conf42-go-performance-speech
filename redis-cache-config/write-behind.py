# Gears Recipe for a single write behind

# import redis gears & mongo db libs
from rgsync import RGJSONWriteBehind
from rgsync.Connectors import MongoConnection, MongoConnector

# change mongodb connection
# MongoConnection(user, password, host, authSource (optional), fullConnectionUrl (optional) )
# connection = MongoConnection('ADMIN_USER','ADMIN_PASSWORD','ADMIN_HOST', "admin")
connection = MongoConnection("", "", "", "", "mongodb://mongodb_server:27017/store")

# change MongoDB database
db = 'store'

# change MongoDB collection & it's primary key
userConnector = MongoConnector(connection, db, 'transactions', 'transactionID')

# change redis keys with prefix that must be synced with mongodb collection
RGJSONWriteBehind(GB, keysPrefix='TransactionEntity',
                  connector=userConnector, name='TransactionsWriteBehind',
                  version='99.99.99')
