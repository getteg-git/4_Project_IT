import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import api from "../../../api/axios";
import "./CreateRepair.css";


function CreateRepair() {

  const [locations, setLocations] = useState([]);
  const [problemTypes, setProblemTypes] = useState([]);

  const [locationId, setLocationId] = useState("");
  const [problemTypeId, setProblemTypeId] = useState("");
  const [description, setDescription] = useState("");

  const [images, setImages] = useState([]);
  const navigate = useNavigate();

  useEffect(() => {
    fetchDropdownData();
  }, []);

  const fetchDropdownData = async () => {
    try {

      const locRes = await api.get("/locations");
      const probRes = await api.get("/problem-types");

      setLocations(locRes.data);
      setProblemTypes(probRes.data);

    } catch (err) {
      console.error(err);
    }
  };

  const handleSubmit = async (e) => {
  e.preventDefault();

  try {

    const formData = new FormData();

    formData.append("user_id", 1);
    formData.append("location_id", locationId);
    formData.append("problem_type_id", problemTypeId);
    formData.append("description", description);

    images.forEach((img) => {
      formData.append("images", img);
    });

    await api.post("/repairs", formData, {
      headers: {
        "Content-Type": "multipart/form-data",
      },
    });

    navigate("/user/home");

  } catch (err) {

    console.error(err);

  }
};

  const handleImageChange = (e) => {

    const files = Array.from(e.target.files);

    setImages((prev) => [...prev, ...files]);

  };

  const removeImage = (index) => {

    setImages(images.filter((_, i) => i !== index));

  };

  return (

    <div className="repair-container">

      <h2>แจ้งซ่อม</h2>

      <form onSubmit={handleSubmit}>

        <label>แนบรูปภาพ</label>

        <input
          type="file"
          multiple
          accept="image/*"
          onChange={handleImageChange}
        />

        <div className="preview-container">

          {images.map((img, index) => (

            <div key={index} className="preview-item">

              <img
                src={URL.createObjectURL(img)}
                alt="preview"
              />

              <button
                type="button"
                onClick={() => removeImage(index)}
              >
                ลบ
              </button>

            </div>

          ))}

        </div>

        <label>สถานที่</label>

        <select
          value={locationId}
          onChange={(e) => setLocationId(e.target.value)}
        >
          <option value="">-- เลือกสถานที่ --</option>

          {locations.map((loc) => (
            <option key={loc.id} value={loc.id}>
              {loc.name}
            </option>
          ))}

        </select>


        <label>ประเภทปัญหา</label>

        <select
          value={problemTypeId}
          onChange={(e) => setProblemTypeId(e.target.value)}
        >

          <option value="">-- เลือกประเภทปัญหา --</option>

          {problemTypes.map((p) => (
            <option key={p.id} value={p.id}>
              {p.name}
            </option>
          ))}

        </select>


        <label>รายละเอียด</label>

        <textarea
          rows="5"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
        />


        <button type="submit">
          แจ้งซ่อม
        </button>

      </form>

    </div>

  );
}

export default CreateRepair;