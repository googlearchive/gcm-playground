# init.sh
# Install dependencies for the web server to run.
# Make sure you have the following installed:
#   - Node.js, npm
#   - git
#   - Docker
#   - Docker Compose


echo "==> Working in web/"
cd web/

echo "==> Install gulp"
if hash gulp 2>/dev/null; then
  echo "gulp installed."
else
    sudo npm install -g gulp
fi

echo "==> Install bower"
if hash gulp 2>/dev/null; then
  echo "bower installed."
else
    sudo npm install -g bower
fi

echo "==> Install npm dependencies"
npm install

echo "==> Install bower dependencies"
bower install --allow-root

echo "==> Running gulp for building"
gulp

cd ..

echo "==> Starting Docker containers"
docker-compose up
