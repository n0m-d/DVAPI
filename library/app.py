from flask import Flask, request, jsonify
from lxml import etree

app = Flask(__name__)

DATA_FILE = "library.xml"


def load_tree():
    return etree.parse(DATA_FILE)


@app.route("/")
def index():
    return jsonify({
        "name": "Library API",
        "auth": "none required",
        "routes": {
            "GET /api/authors": "list all authors",
            "GET /api/authors/<id>": "get one author",
            "GET /api/books": "list all books",
            "GET /api/books/<id>": "get one book",
            "GET /api/books?title=<query>": "search books by title FAST !!!",
        }
    })


@app.route("/api/authors", methods=["GET"])
def list_authors():
    tree = load_tree()
    authors = []
    for a in tree.xpath("//author"):
        authors.append({
            "id": a.get("id"),
            "name": a.get("name"),
            "country": a.get("country"),
        })
    return jsonify(authors)


@app.route("/api/authors/<author_id>", methods=["GET"])
def get_author(author_id):
    tree = load_tree()
    query ="//author[@id=$id]"
    result = tree.xpath(query,id=author_id)
    if not result:
        return jsonify({"error": "author not found"}), 404
    a = result[0]
    return jsonify({
        "id": a.get("id"),
        "name": a.get("name"),
        "country": a.get("country"),
        "books": [b.get("title") for b in a.xpath("./book")],
    })


@app.route("/api/books/<book_id>", methods=["GET"])
def get_book(book_id):
    tree = load_tree()
    query= "//book[@id=$id]"
    result = tree.xpath(query,id=book_id)
    if not result:
        return jsonify({"error": "book not found"}), 404
    b = result[0]
    author = b.getparent()
    return jsonify({
        "id": b.get("id"),
        "title": b.get("title"),
        "year": b.get("year"),
        "genre": b.get("genre"),
        "author": author.get("name"),
    })


@app.route("/api/books", methods=["GET"])
def books():
    tree = load_tree()
    title = request.args.get("title")

    def mask(text):
        return text if len(text) <= 2 else text[0] + "#" * (len(text) - 2) + text[-1]

    def format(book, detailed=False):
        data = {
            "id": book.get("id"),
            "title": book.get("title"),
            "author": book.getparent().get("name"),
        }

        if detailed:
            data.update({
                "year": book.get("year"),
                "genre": book.get("genre"),
            })

        if (bm := book.find("bookmark")) is not None and bm.text:
            data["bookmark"] = mask(bm.text)

        return data

    try:
        matches = tree.xpath(f"//book[@title='{title}']") if title else tree.xpath("//book")
    except etree.XPathEvalError:
        return jsonify({"found": False, "books": []}), 404

    books = [format(b, detailed=not title) for b in matches]

    if title:
        return jsonify({"found": bool(books), "books": books}), 200 if books else 404

    return jsonify({"found": bool(books), "books": books}), 200

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5000, debug=True)
