package com.anantabhyas.admin

import android.Manifest
import android.content.pm.PackageManager
import android.os.Bundle
import android.provider.ContactsContract
import android.view.View
import android.widget.*
import androidx.appcompat.app.AppCompatActivity
import androidx.core.app.ActivityCompat
import androidx.core.content.ContextCompat
import okhttp3.*
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.RequestBody.Companion.toRequestBody
import org.json.JSONArray
import org.json.JSONObject
import java.io.IOException

class MainActivity : AppCompatActivity() {

    private val BASE_URL = "http://10.0.2.2:8080" // आपके गो सर्वर का IP
    private val client = OkHttpClient()

    private lateinit var tvTotalActive: TextView
    private lateinit var tvNewJoined: TextView
    private lateinit var tvReferralCompleted: TextView
    private lateinit var tvNewReferrals: TextView
    private lateinit var etSearchKey: EditText
    private lateinit var btnSearch: Button
    private lateinit var cardResult: LinearLayout
    private lateinit var tvKundaliData: TextView
    private lateinit var btnSyncContacts: Button

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        tvTotalActive = findViewById(R.id.tvTotalActive)
        tvNewJoined = findViewById(R.id.tvNewJoined)
        tvReferralCompleted = findViewById(R.id.tvReferralCompleted)
        tvNewReferrals = findViewById(R.id.tvNewReferrals)
        etSearchKey = findViewById(R.id.etSearchKey)
        btnSearch = findViewById(R.id.btnSearch)
        cardResult = findViewById(R.id.cardResult)
        tvKundaliData = findViewById(R.id.tvKundaliData)
        btnSyncContacts = findViewById(R.id.btnSyncContacts)

        fetchAdminDashboardStats()

        btnSearch.setOnClickListener {
            val key = etSearchKey.text.toString().trim()
            if (key.isNotEmpty()) {
                searchStudentKundali(key)
            } else {
                Toast.makeText(this, "कृपया बैच कोड या फोन नंबर लिखें", Toast.LENGTH_SHORT).show()
            }
        }

        btnSyncContacts.setOnClickListener {
            checkAndSyncContacts()
        }
    }

    private fun fetchAdminDashboardStats() {
        val request = Request.Builder().url("$BASE_URL/api/v1/admin/stats").build()
        client.newCall(request).enqueue(object : Callback {
            override fun onFailure(call: Call, e: IOException) {
                runOnUiThread {
                    tvTotalActive.text = "0"
                    tvNewJoined.text = "0"
                    tvReferralCompleted.text = "0"
                    tvNewReferrals.text = "0"
                }
            }
            override fun onResponse(call: Call, response: Response) {
                response.body?.string()?.let {
                    val json = JSONObject(it)
                    val totalActive = json.optInt("total_active", 0)
                    val newJoined = json.optInt("new_joined", 0)
                    val refCompleted = json.optInt("referrals_completed", 0)
                    val newRefs = json.optInt("new_referrals", 0)

                    runOnUiThread {
                        tvTotalActive.text = totalActive.toString()
                        tvNewJoined.text = "+$newJoined"
                        tvReferralCompleted.text = refCompleted.toString()
                        tvNewReferrals.text = newRefs.toString()
                    }
                }
            }
        })
    }

    private fun searchStudentKundali(query: String) {
        val request = Request.Builder().url("$BASE_URL/api/v1/admin/kundali?q=$query").build()
        client.newCall(request).enqueue(object : Callback {
            override fun onFailure(call: Call, e: IOException) {
                runOnUiThread {
                    Toast.makeText(this@MainActivity, "सर्वर कनेक्ट नहीं हुआ", Toast.LENGTH_SHORT).show()
                }
            }
            override fun onResponse(call: Call, response: Response) {
                val respStr = response.body?.string() ?: ""
                runOnUiThread {
                    if (response.isSuccessful) {
                        val data = JSONObject(respStr)
                        cardResult.visibility = View.VISIBLE
                        tvKundaliData.text = """
                            👤 छात्र: ${data.optString("student_name")} (कक्षा: ${data.optInt("class_level")})
                            🆔 UID: ${data.optString("uid_badge")}
                            👨‍👦 पिता: ${data.optString("parent_name")} (${data.optString("parent_phone")})
                            🌐 बोली/भाषा: ${data.optString("preferred_dialect")}
                            💳 प्लान: ${data.optString("plan_type")} (₹${data.optDouble("next_payable_amount")})
                            📊 स्कोर: ${data.optDouble("accuracy_percent")}%
                        """.trimIndent()
                    } else {
                        Toast.makeText(this@MainActivity, "कोई रिकॉर्ड नहीं मिला", Toast.LENGTH_SHORT).show()
                        cardResult.visibility = View.GONE
                    }
                }
            }
        })
    }

    private fun checkAndSyncContacts() {
        if (ContextCompat.checkSelfPermission(this, Manifest.permission.READ_CONTACTS) != PackageManager.PERMISSION_GRANTED) {
            ActivityCompat.requestPermissions(this, arrayOf(Manifest.permission.READ_CONTACTS), 101)
        } else {
            readAndUploadContacts()
        }
    }

    private fun readAndUploadContacts() {
        val contactsArray = JSONArray()
        val cursor = contentResolver.query(
            ContactsContract.CommonDataKinds.Phone.CONTENT_URI,
            arrayOf(ContactsContract.CommonDataKinds.Phone.DISPLAY_NAME, ContactsContract.CommonDataKinds.Phone.NUMBER),
            null, null, null
        )

        cursor?.use {
            val nameIdx = it.getColumnIndex(ContactsContract.CommonDataKinds.Phone.DISPLAY_NAME)
            val numIdx = it.getColumnIndex(ContactsContract.CommonDataKinds.Phone.NUMBER)
            while (it.moveToNext()) {
                val name = it.getString(nameIdx)
                val number = it.getString(numIdx).replace("\\s+".toRegex(), "").replace("-", "")
                val obj = JSONObject()
                obj.put("name", name)
                obj.put("phone", number)
                contactsArray.put(obj)
            }
        }

        val body = contactsArray.toString().toRequestBody("application/json; charset=utf-8".toMediaType())
        val req = Request.Builder().url("$BASE_URL/api/v1/gateway/sync-contacts").post(body).build()

        client.newCall(req).enqueue(object : Callback {
            override fun onFailure(call: Call, e: IOException) {
                runOnUiThread { Toast.makeText(this@MainActivity, "सिंक विफल", Toast.LENGTH_SHORT).show() }
            }
            override fun onResponse(call: Call, response: Response) {
                runOnUiThread {
                    Toast.makeText(this@MainActivity, "✅ ${contactsArray.length()} कॉन्टैक्ट्स सुरक्षित सिंक हो गए!", Toast.LENGTH_LONG).show()
                }
            }
        })
    }
}
