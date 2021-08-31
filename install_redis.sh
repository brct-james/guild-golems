#!/bin/sh

echo "Using rejson rather than pure redis"
docker run -p 6380:6379 --name redis-redisjson redislabs/rejson:latest


# # In order to compile Redis follow these simple steps:
# cd ~
# wget http://download.redis.io/redis-stable.tar.gz
# tar xvzf redis-stable.tar.gz
# cd redis-stable
# make

# # It is a good idea to copy both the Redis server and the command line interface into the proper places, either manually using the following commands:
# # sudo cp src/redis-server /usr/local/bin/
# # sudo cp src/redis-cli /usr/local/bin/
# sudo make install

# # Create a directory in which to store your Redis config files and your data:
# sudo mkdir /etc/redis
# sudo mkdir /var/redis

# #Copy the init script that you'll find in the Redis distribution under the utils directory into /etc/init.d. We suggest calling it with the name of the port where you are running this instance of Redis. For example
# sudo cp utils/redis_init_script /etc/init.d/redis_6379

# #Edit the init script
# # sudo vi /etc/init.d/redis_6379

# #Copy the template configuration file you'll find in the root directory of the Redis distribution into /etc/redis/ using the port number as name, for instance
# sudo cp redis.conf /etc/redis/6379.conf

# #Create a directory inside /var/redis that will work as data and working directory for this Redis instance
# sudo mkdir /var/redis/6379

# # # # Edit the configuration file, making sure to perform the following changes:
# # Set daemonize to yes (by default it is set to no).
# # Set the pidfile to /var/run/redis_6379.pid (modify the port if needed).
# # Change the port accordingly. In our example it is not needed as the default port is already 6379.
# # Set your preferred loglevel.
# # Set the logfile to /var/log/redis_6379.log
# # Set the dir to /var/redis/6379 (very important step!)
# sudo vim /etc/redis/6379.conf

# #Finally add the new Redis init script to all the default runlevels using the following command:
# sudo update-rc.d redis_6379 defaults

# #You are done! Now you can try running your instance with:
# sudo /etc/init.d/redis_6379 start

# Make sure that everything is working as expected:

# # Try pinging your instance with redis-cli.
# # Do a test save with redis-cli save and check that the dump file is correctly stored into /var/redis/6379/ (you should find a file called dump.rdb).
# # Check that your Redis instance is correctly logging in the log file.
# # If it's a new machine where you can try it without problems make sure that after a reboot everything is still working.