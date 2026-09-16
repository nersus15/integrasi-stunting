// Pencarian lintas halaman. Markdown tiap halaman diambil lewat ?doc=<n>&raw=1
// sekali saja saat pencarian pertama, lalu disimpan di memori.
(function () {
  "use strict";

  var halaman = [];
  try {
    halaman = JSON.parse(document.getElementById("daftar-halaman").textContent);
  } catch (e) {
    return;
  }

  var kotak = document.getElementById("cari");
  var hasil = document.getElementById("hasil-cari");
  if (!kotak || !hasil) return;

  var isi = null;
  var sedangMuat = false;

  function muat() {
    if (isi || sedangMuat) return Promise.resolve();
    sedangMuat = true;
    return Promise.all(
      halaman.map(function (h) {
        return fetch("?doc=" + encodeURIComponent(h.nama) + "&raw=1")
          .then(function (r) { return r.ok ? r.text() : ""; })
          .then(function (t) { return { nama: h.nama, judul: h.judul, teks: t }; })
          .catch(function () { return { nama: h.nama, judul: h.judul, teks: "" }; });
      })
    ).then(function (semua) {
      isi = semua;
      sedangMuat = false;
    });
  }

  // judul bagian terdekat di atas posisi kecocokan, supaya hasilnya bisa ditaut
  function bagianTerdekat(teks, posisi) {
    var sebelum = teks.slice(0, posisi);
    var baris = sebelum.split("\n");
    for (var i = baris.length - 1; i >= 0; i--) {
      var m = /^#{1,4}\s+(.*)$/.exec(baris[i]);
      if (m) return m[1].replace(/`/g, "");
    }
    return "";
  }

  function slug(judul) {
    return judul.toLowerCase().trim()
      .replace(/[^\w\s-]/g, "")
      .replace(/\s+/g, "-");
  }

  function cuplik(teks, posisi, panjang) {
    var awal = Math.max(0, posisi - 60);
    var akhir = Math.min(teks.length, posisi + panjang + 60);
    var s = teks.slice(awal, akhir).replace(/\n+/g, " ").trim();
    return (awal > 0 ? "…" : "") + s + (akhir < teks.length ? "…" : "");
  }

  function cari(kata) {
    var q = kata.toLowerCase();
    var keluar = [];

    isi.forEach(function (h) {
      var rendah = h.teks.toLowerCase();
      var dari = 0;
      var n = 0;
      while (n < 5) {
        var p = rendah.indexOf(q, dari);
        if (p < 0) break;
        var bagian = bagianTerdekat(h.teks, p);
        keluar.push({
          halaman: h.judul,
          nama: h.nama,
          bagian: bagian,
          cuplikan: cuplik(h.teks, p, q.length),
        });
        dari = p + q.length;
        n++;
      }
    });
    return keluar;
  }

  function tulis(daftar, kata) {
    hasil.replaceChildren();
    if (!kata) {
      hasil.hidden = true;
      return;
    }
    hasil.hidden = false;

    if (!daftar.length) {
      var kosong = document.createElement("p");
      kosong.className = "cari-kosong";
      kosong.textContent = 'Tidak ada yang cocok dengan "' + kata + '".';
      hasil.appendChild(kosong);
      return;
    }

    var ringkas = document.createElement("p");
    ringkas.className = "cari-ringkas";
    ringkas.textContent = daftar.length + " kecocokan";
    hasil.appendChild(ringkas);

    daftar.forEach(function (r) {
      var a = document.createElement("a");
      a.className = "cari-item";
      a.href = "?doc=" + encodeURIComponent(r.nama) + (r.bagian ? "#" + slug(r.bagian) : "");

      var atas = document.createElement("span");
      atas.className = "cari-lokasi";
      atas.textContent = r.halaman + (r.bagian ? " › " + r.bagian : "");
      a.appendChild(atas);

      var bawah = document.createElement("span");
      bawah.className = "cari-cuplikan";
      var i = r.cuplikan.toLowerCase().indexOf(kata.toLowerCase());
      if (i < 0) {
        bawah.textContent = r.cuplikan;
      } else {
        bawah.appendChild(document.createTextNode(r.cuplikan.slice(0, i)));
        var tandai = document.createElement("mark");
        tandai.textContent = r.cuplikan.slice(i, i + kata.length);
        bawah.appendChild(tandai);
        bawah.appendChild(document.createTextNode(r.cuplikan.slice(i + kata.length)));
      }
      a.appendChild(bawah);
      hasil.appendChild(a);
    });
  }

  var jeda;
  kotak.addEventListener("input", function () {
    clearTimeout(jeda);
    var kata = kotak.value.trim();
    jeda = setTimeout(function () {
      if (kata.length < 2) {
        tulis([], "");
        return;
      }
      muat().then(function () {
        tulis(cari(kata), kata);
      });
    }, 150);
  });

  kotak.addEventListener("keydown", function (e) {
    if (e.key === "Escape") {
      kotak.value = "";
      tulis([], "");
      kotak.blur();
    }
  });

  document.addEventListener("keydown", function (e) {
    if (e.key === "/" && document.activeElement !== kotak) {
      e.preventDefault();
      kotak.focus();
    }
  });
})();
