Composable Tutorial Example
===========================

This file uses composable tutorials with selected-content blocks.

.. composable-tutorial::
   :options: language=python; interface=driver

   Introduction to the tutorial.

   .. selected-content::
      :selections: python

      Python-specific content:

      .. code-block:: python

         # This is Python code in a selected-content block
         result = collection.find_one()

      .. include:: /includes/python-example.rst

   .. selected-content::
      :selections: nodejs

      Node.js-specific content:

      .. code-block:: javascript

         // This is Node.js code in a selected-content block
         const result = await collection.findOne();

      .. include:: /includes/nodejs-example.rst

