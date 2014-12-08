package com.cherami.cherami;

import android.app.ProgressDialog;
import android.content.Context;
import android.content.Intent;
import android.os.Bundle;
import android.app.Fragment;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.ListView;
import android.widget.TextView;
import android.widget.Toast;

import com.loopj.android.http.AsyncHttpClient;
import com.loopj.android.http.AsyncHttpResponseHandler;

import org.apache.http.Header;
import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;

public class Profile extends Fragment {
    private ListView messageList;
    TextView textElement;
    Context context;
    ProgressDialog dialog;
    FeedAdapter adapter;

    public Profile() {

    }

    @Override
    public void onCreate(Bundle savedInstanceState) {
        this.context = getActivity().getApplicationContext();
        super.onCreate(savedInstanceState);
    }

    @Override
    public View onCreateView(LayoutInflater inflater, ViewGroup container, Bundle savedInstanceState) {
        String username = ApiHelper.getUsername(context);
        View rootView = inflater.inflate(R.layout.fragment_profile, container, false);
        getProfileFeed(rootView);
        textElement = (TextView) rootView.findViewById(R.id.profileHandle);
        textElement.setText(username);

        return rootView;
    }

    public void getProfileFeed(final View view) {
        AsyncHttpClient client = new AsyncHttpClient();
        String token = ApiHelper.getSessionToken(context);

        client.addHeader("Authorization", token);
        client.get(context,
                   ApiHelper.getLocalUrlForApi(getResources()) + "messages",
                   new AsyncHttpResponseHandler() {

            @Override
            public void onStart() {
                dialog = ProgressDialog.show(getActivity(), "",
                        "Loading. Please wait...", true);
            }

            @Override
            public void onSuccess(int statusCode, Header[] headers, byte[] responseBody) {
                JSONArray responseText;

                try {
                    responseText = new JSONArray(new String(responseBody));
                    FeedItem feed_data[] = new FeedItem[responseText.length()];

                    for (int x = 0; x < responseText.length(); x++) {
                        feed_data[x] = new FeedItem(new JSONObject(responseText.get(x).toString()));
                    }

                    final FeedAdapter adapter = new FeedAdapter(getActivity(),
                            R.layout.feed_item_row, feed_data);

                    messageList = (ListView) view.findViewById(R.id.profileFeed);
                    messageList.setAdapter(adapter);

                } catch (JSONException j) {

                }
                dialog.dismiss();

            }

            @Override
            public void onFailure(int statusCode, Header[] headers, byte[] errorResponse, Throwable error) {
                dialog.dismiss();
                String responseText = null;

                try {
                    if (!NetworkCheck.isConnected(errorResponse)) {
                        NetworkCheck.displayNetworkErrorModal(getActivity());

                    } else {
                        responseText = new JSONObject(new String(errorResponse)).getString("reason");
                        Toast toast = Toast.makeText(getActivity().getApplicationContext(), responseText, Toast.LENGTH_LONG);
                        toast.show();
                    }
                } catch (JSONException j) {

                }

            }
        });
    }

    public String processDate(String date){
        return date.substring(0, date.lastIndexOf("T"));
    }

    public void showLogin(View view) {
        Intent intent = new Intent(getActivity(), LoginActivity.class);
        startActivity(intent);
    }
}
