Driver Tabs Example
===================

This file contains code examples within driver tabs.

.. tabs-drivers::

   .. tab::
      :tabid: python

      Python driver example:

      .. code-block:: python

         from pymongo import MongoClient
         client = MongoClient()

   .. tab::
      :tabid: nodejs

      Node.js driver example:

      .. code-block:: javascript

         const { MongoClient } = require('mongodb');
         const client = new MongoClient(uri);

   .. tab::
      :tabid: java-sync

      Java driver example:

      .. code-block:: java

         MongoClient client = MongoClients.create();

