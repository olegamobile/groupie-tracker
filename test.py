import requests

# Fetch data from the API again to generate all 52 cards
response = requests.get("https://groupietrackers.herokuapp.com/api/artists")
artists_data = response.json()

# Begin building the HTML content
html_content = """
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Groupie tracker</title>
    <!-- Bootstrap CSS -->
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
</head>
<body>
    <div class="container my-4">
        <div class="row g-4">
"""

# Loop through the first 52 artists and add cards for each artist
for artist in artists_data[:52]:
    name = artist.get("name", "Unknown Artist")
    image = artist.get("image", "https://via.placeholder.com/150")
    members = ", ".join(artist.get("members", [])) or "Unknown Members"
    first_album = artist.get("firstAlbum", "Unknown Date")

    html_content += f"""
            <div class="col-12 col-sm-6 col-md-4 col-lg-3">
                <div class="card h-100">
                    <img src="{image}" class="card-img-top" alt="{name}">
                    <div class="card-body d-flex flex-column">
                        <h5 class="card-title">{name}</h5>
                        <p class="card-text"><strong>Members:</strong> {members}</p>
                        <p class="card-text"><strong>First Album:</strong> {first_album}</p>
                        <a href="#" class="btn btn-primary mt-auto">Details</a>
                    </div>
                </div>
            </div>
    """

# Close HTML tags
html_content += """
        </div>
    </div>
    <!-- Bootstrap JS -->
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
</body>
</html>
"""

# Save the generated HTML content to a file
file_path_52_cards = "Artists_52_Cards.html"
with open(file_path_52_cards, "w") as file:
    file.write(html_content)

file_path_52_cards
