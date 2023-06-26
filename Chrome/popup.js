$(function () {
  let host = 'http://127.0.0.1:5000'
  $('#rolad_waller').click(function () {
    $.ajax({
      url: "http://localhost:8080/wallet",
      type: "POST",
      success: function (response) {
        $("#inputPublic").val(response["public_key"]);
        $("#inputPrivateKey").val(response["private_key"]);
        $("#inputAddress").val(response["blockchain_address"]);
        $("#account").text(response["blockchain_address"]);
        console.info(response);
      },
      error: function (error) {
        console.error(error);
      },
    })
  })
  $('#select_vale').change(function () {

    host = $('#select_vale').val()
    alert("服务为" + host)
  })

  $('#getBlance').click(function () {


    let account = $("#account").text();

    let requestData = {
      blockchain_address: account
    };

    $.ajax({
      url: "http://localhost:8080/wallet/amount",
      type: "POST",
      data: JSON.stringify(requestData),
      contentType: "application/json",
      success: function (response) {
        $("#blance").text(response.amount)
      },
      error: function (error) {
        console.error(error);
      },
    });
  });


  $('#load_waller').click(function () {
    let privatekey = $("#inputPrivateKey").val();
    $.ajax({
      url: "http://localhost:8080/walletByPrivatekey",
      type: "POST",
      data: {
        private: privatekey
      },
      success: function (response) {
        $("#inputPublic").val(response["public_key"]);
        $("#account").text(response["blockchain_address"]);
        $("#inputPrivateKey").val(response["private_key"]);
        $("#inputAddress").val(response["blockchain_address"]);
      },
      error: function (error) {
        console.error(error);
      },
    })
  })


  $('#buttonSubmit').click(function () {
    let sender_blockchain_address = $("#inputAddress").val();
    let sender_private_key = $("#inputPrivateKey").val();
    let sender_public_key = $("#inputPublic").val();
    let recipient_blockchain_address = $("#inputReceiveAddress").val();
    let value = $("#inputAmount").val();

    let requestData = {
      sender_blockchain_address: sender_blockchain_address,
      sender_private_key: sender_private_key,
      sender_public_key: sender_public_key,
      recipient_blockchain_address: recipient_blockchain_address,
      value: value
    };

    $.ajax({
      url: "http://localhost:8080/transaction",
      type: "POST",
      data: JSON.stringify(requestData),
      contentType: "application/json",
      success: function (response) {
        console.log(response);
        if (response.message == "fail") {
          alert("转账失败")
          return
        }
        alert("转账成功")

      },
      error: function (error) {
        console.error(error);
      },
    });
  });


  document.querySelectorAll('.clickable')[0].addEventListener('click', handleClick)
  function handleClick() {
    document.getElementById('nav-contact').innerHTML = ""
    $.ajax({
      url: host,
      type: "GET",
      success: function (response) {
        renderData(response)
      },
      error: function (error) {
        document.getElementById('nav-contact').innerHTML = ""
        alert("服务找不到")
      },
    })

  }


  const renderData = (data) => {
    let html = '';
    const dataContainer = document.getElementById('nav-contact');
    data.chain.forEach((block, index) => {
      // 选择颜色类名
      const colorClass = index % 2 === 0 ? 'card-primary' : 'card-secondary';

      html += `
        <div class="card my-4 ${colorClass}">
          <div class="card-header">
            Block Number: ${block.number}
          </div>
          <div class="card-body">
            <p class="mb-2"><strong>Timestamp:</strong> ${block.timestamp}</p>
            <p class="mb-2"><strong>Nonce:</strong> ${block.nonce}</p>
            <p class="mb-2"><strong>Difficulty:</strong> ${block.difficulty}</p>
            <p class="mb-2"><strong>Hash:</strong> ${block.hash}</p>
            <p class="mb-2"><strong>Previous Hash:</strong> ${block.previous_hash}</p>
            <p class="mb-2"><strong>Transaction Size:</strong> ${block.txSize}</p>
            <ul class="list-group">
              ${block.transactions.map((transaction) => `
                <li class="list-group-item transaction-item">
                  <p class="mb-1"><strong>Sender:</strong> ${transaction.sender_blockchain_address}</p>
                  <p class="mb-1"><strong>Recipient:</strong> ${transaction.recipient_blockchain_address}</p>
                  <p class="mb-1"><strong>Value:</strong> ${transaction.value}</p>
                  <p class="mb-1"><strong>Timestamp:</strong> ${transaction.timestamp}</p>
                  <p class="mb-1"><strong>Transaction Hash:</strong> ${transaction.transactionHash}</p>
                </li>
              `).join('')}
            </ul>
          </div>
        </div>
      `;
    });

    // 将渲染后的HTML插入到数据容器中
    dataContainer.innerHTML = html;
  };









  const searchForm = document.getElementById('search-form');
  const searchInput = document.getElementById('search-input');
  const searchType = document.getElementById('search-type');
  const searchResults = document.getElementById('search-results');
  const errorMessageContainer = document.getElementById('error-message');

  searchForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    const query = searchInput.value;
    const type = searchType.value;
    try {
      const results = await searchBlockchain(query, type);
      displaySearchResults(results, type);
    } catch (error) {
      displayErrorMessage(error);
    }
  });

  function displayErrorMessage(message) {
    searchResults.innerHTML = ''
    errorMessageContainer.textContent = message;
  }

  function searchBlockchain(query, type) {
    return new Promise((resolve, reject) => {
      if (type === 'blockNumber') {
        let requestData = {
          blockNumber: query
        };
        $.ajax({
          url: host + "/GetBlockByNumber",
          type: "POST",
          data: JSON.stringify(requestData),
          contentType: "application/json",
          success: function (response) {
            resolve(response);
          },
          error: function (error) {
            reject(error.responseText);
          },
        });
      } else if (type === 'blockHash') {
        let requestData = {
          blockHash: query
        };
        $.ajax({
          url: host + "/GetBlockByHash",
          type: "POST",
          data: JSON.stringify(requestData),
          contentType: "application/json",
          success: function (response) {
            resolve(response);
          },
          error: function (error) {
            reject(error.responseText);
          },
        });

      } else if (type === 'transactionHash') {

        let requestData = {
          transactionHash: query
        };
        $.ajax({
          url: host + "/GetTransactionByHash",
          type: "POST",
          data: JSON.stringify(requestData),
          contentType: "application/json",
          success: function (response) {
            resolve(response);
          },
          error: function (error) {
            reject(error.responseText);
          },
        });
      }
    });
  }
  function displaySearchResults(results, type) {
    errorMessageContainer.textContent = ""
    let html = '';
    if (type === 'blockNumber') {
      html += `
        <div class="card my-4">
          <div class="card-header">
            Block Number: ${results.number}
          </div>
          <div class="card-body">
            <p class="mb-2"><strong>Timestamp:</strong> ${results.timestamp}</p>
            <p class="mb-2"><strong>Nonce:</strong> ${results.nonce}</p>
            <p class="mb-2"><strong>Difficulty:</strong> ${results.difficulty}</p>
            <p class="mb-2"><strong>Hash:</strong> ${results.hash}</p>
            <p class="mb-2"><strong>Previous Hash:</strong> ${results.previous_hash}</p>
            <p class="mb-2"><strong>Transaction Size:</strong> ${results.txSize}</p>
            <ul class="list-group">
              ${results.transactions.map((transaction) => `
                <li class="list-group-item">
                  <p class="mb-1"><strong>Sender:</strong> ${transaction.sender_blockchain_address}</p>
                  <p class="mb-1"><strong>Recipient:</strong> ${transaction.recipient_blockchain_address}</p>
                  <p class="mb-1"><strong>Value:</strong> ${transaction.value}</p>
                  <p class="mb-1"><strong>Timestamp:</strong> ${transaction.timestamp}</p>
                  <p class="mb-1"><strong>Transaction Hash:</strong> ${transaction.transactionHash}</p>
                </li>
              `).join('')}
            </ul>
          </div>
        </div>
      `;
    } else if (type === 'blockHash') {
      html += `
        <div class="card my-4">
          <div class="card-header">
            Block Hash: ${results.hash}
          </div>
          <div class="card-body">
            <p class="mb-2"><strong>Number:</strong> ${results.number}</p>
            <p class="mb-2"><strong>Timestamp:</strong> ${results.timestamp}</p>
            <p class="mb-2"><strong>Nonce:</strong> ${results.nonce}</p>
            <p class="mb-2"><strong>Difficulty:</strong> ${results.difficulty}</p>
            <p class="mb-2"><strong>Previous Hash:</strong> ${results.previous_hash}</p>
            <p class="mb-2"><strong>Transaction Size:</strong> ${results.txSize}</p>
            <ul class="list-group">
              ${results.transactions.map((transaction) => `
                <li class="list-group-item">
                  <p class="mb-1"><strong>Sender:</strong> ${transaction.sender_blockchain_address}</p>
                  <p class="mb-1"><strong>Recipient:</strong> ${transaction.recipient_blockchain_address}</p>
                  <p class="mb-1"><strong>Value:</strong> ${transaction.value}</p>
                  <p class="mb-1"><strong>Timestamp:</strong> ${transaction.timestamp}</p>
                  <p class="mb-1"><strong>Transaction Hash:</strong> ${transaction.transactionHash}</p>
                </li>
              `).join('')}
            </ul>
          </div>
        </div>
      `;
    } else if (type === 'transactionHash') {
      html += `
        <div class="card my-4">
          <div class="card-header">
            Transaction Hash: ${results.transactionHash}
          </div>
          <div class="card-body">
            <p class="mb-2"><strong>Sender:</strong> ${results.sender_blockchain_address}</p>
            <p class="mb-2"><strong>Recipient:</strong> ${results.recipient_blockchain_address}</p>
            <p class="mb-2"><strong>Value:</strong> ${results.value}</p>
            <p class="mb-2"><strong>Timestamp:</strong> ${results.timestamp}</p>
            <p class="mb-2"><strong>Transaction Hash:</strong> ${results.transactionHash}</p>
          </div>
        </div>
      `;
    } else {
      html = '<p>No results found.</p>';
    }

    searchResults.innerHTML = html;
  }





});






var storage = chrome.storage.sync;
// 假设你有一个保存按钮的元素，其 ID 为 saveButton
document.getElementById('load_waller').addEventListener('click', function () {
  // 获取用户输入的数据
  var inputData = document.getElementById('inputPrivateKey').value;

  // 将数据保存到存储中
  storage.set({ 'data': inputData }, function () {
    console.log('数据已保存');
  });
});

document.getElementById('rolad_waller').addEventListener('click', function () {
  // 获取用户输入的数据
  var inputData = document.getElementById('inputPrivateKey').value;

  // 将数据保存到存储中
  storage.set({ 'data': inputData }, function () {
    console.log('数据已保存');
  });
});

document.addEventListener('DOMContentLoaded', function () {
  // 从存储中检索数据
  storage.get('data', function (result) {
    // 检查是否存在之前保存的数据
    if (result.data) {
      // 将数据应用到用户界面
      document.getElementById('inputPrivateKey').value = result.data;
    }
  });
});
