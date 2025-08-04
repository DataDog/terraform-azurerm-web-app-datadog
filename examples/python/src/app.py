import os
import logging
from ddtrace.trace import tracer
from flask import Flask

app = Flask(__name__)

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

@app.route("/")
@tracer.wrap()
def hello_world():
    logger.info("Hello Datadog logger using Python!")
    return "Hello World from Python!"

if __name__ == "__main__":
    app.run(debug=True, host="0.0.0.0", port=int(os.environ.get("PORT", 8080)))
