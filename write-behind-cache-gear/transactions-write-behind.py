# Gears Recipe for a single write behind

# gears redis steps to execute this script
# gears-cli import-requirements --host localhost --port 27017 --requirements-path
# redis-cli RG.PYEXECUTE "`cat wbc.py`"

# import redis gears & mongo db libs
from rgsync import RGJSONWriteBehind, RGJSONWriteThrough
from rgsync.Connectors import MongoConnector, MongoConnection

# change mongodb connection (admin)
# mongodb://usrAdmin:passwordAdmin@localhost:27017/dbSpeedMernDemo?authSource=admin
mongoUrl = 'mongodb://localhost:27017'

# MongoConnection(user, password, host, authSource?, fullConnectionUrl?)
connection = MongoConnection('', '', '', '', mongoUrl)

# change MongoDB database
db = 'dbSpeedMernDemo'

# change MongoDB collection & it's primary key
transactionsConnector = MongoConnector(connection, db, 'transactions', 'transactionID')

# change redis keys with prefix that must be synced with mongodb collection
RGJSONWriteBehind(GB,  keysPrefix='TransactionEntity',
                  connector=transactionsConnector, name='TransactionsWriteBehind',
                  version='99.99.99')